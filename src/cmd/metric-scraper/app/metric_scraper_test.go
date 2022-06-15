package app_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"code.cloudfoundry.org/tlsconfig"
	"google.golang.org/grpc/credentials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"google.golang.org/grpc"

	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
	metricshelper "code.cloudfoundry.org/go-metric-registry/testhelpers"
	"code.cloudfoundry.org/system-metrics-scraper/cmd/metric-scraper/app"
	"code.cloudfoundry.org/system-metrics-scraper/internal/testhelper"
)

var _ = Describe("App", func() {
	var (
		spyAgent    *spyAgent
		dnsFilePath string
		scraper     *app.MetricScraper
		cfg         app.Config

		testLogger       = gbytes.NewBuffer()
		leadership       *spyLeadership
		promServer       *promServer
		promAddr         string
		spyMetricsClient *metricshelper.SpyMetricsRegistry
		spyNATSConn      *spyNATSConn

		leadershipTestCerts    = testhelper.GenerateCerts("leadershipCA")
		metronTestCerts        = testhelper.GenerateCerts("loggregatorCA")
		systemMetricsTestCerts = testhelper.GenerateCerts("systemMetricsCA")
	)

	Describe("when configured with a single metrics_url", func() {
		BeforeEach(func() {
			spyAgent = newSpyAgent(metronTestCerts)
			leadership = newSpyLeadership(leadershipTestCerts)

			promServer = newPromServer()
			promServer.start(systemMetricsTestCerts)

			u, err := url.Parse(promServer.url())
			Expect(err).ToNot(HaveOccurred())

			scrapePort, err := strconv.Atoi(u.Port())
			Expect(err).ToNot(HaveOccurred())

			promAddr = u.Hostname()
			dnsFilePath = createDNSFile(promAddr)

			cfg = app.Config{
				ClientKeyPath:          metronTestCerts.Key("metron"),
				ClientCertPath:         metronTestCerts.Cert("metron"),
				CACertPath:             metronTestCerts.CA(),
				MetricsKeyPath:         systemMetricsTestCerts.Key("system-metrics-agent"),
				MetricsCertPath:        systemMetricsTestCerts.Cert("system-metrics-agent"),
				MetricsCACertPath:      systemMetricsTestCerts.CA(),
				LeadershipKeyPath:      leadershipTestCerts.Key("leadership-client"),
				LeadershipCertPath:     leadershipTestCerts.Cert("leadership-client"),
				LeadershipCACertPath:   leadershipTestCerts.CA(),
				MetricsCN:              "system-metrics-agent",
				LoggregatorIngressAddr: spyAgent.addr,
				ScrapeInterval:         100 * time.Millisecond,
				ScrapePort:             scrapePort,
				DefaultSourceID:        "default-id",
				DNSFile:                dnsFilePath,
				LeadershipServerAddr:   leadership.server.URL,
			}

			spyMetricsClient = metricshelper.NewMetricsRegistry()
		})

		AfterEach(func() {
			scraper.Stop()
			os.RemoveAll(filepath.Dir(dnsFilePath))
		})

		It("scrapes a prometheus endpoint and sends those metrics to a loggregator agent", func() {
			scraper = app.NewMetricScraper(cfg, testLogger, spyMetricsClient, emptyNATSConn{})
			go scraper.Run()

			Eventually(spyAgent.Envelopes).Should(And(
				ContainElement(buildCounter("source-1", "node_timex_pps_calibration_total", promAddr, 1)),
				ContainElement(buildCounter("source-1", "node_timex_pps_error_total", promAddr, 2)),
				ContainElement(buildGauge("source-1", "node_timex_pps_frequency_hertz", promAddr, 3)),
				ContainElement(buildGauge("source-2", "node_timex_pps_jitter_seconds", promAddr, 4)),
				ContainElement(buildCounter("default-id", "node_timex_pps_jitter_total", promAddr, 5)),
			))
		})

		It("does not scrape when leadership server returns 423", func() {
			leadership.setReturnCode(http.StatusLocked)

			scraper = app.NewMetricScraper(cfg, testLogger, spyMetricsClient, emptyNATSConn{})
			go scraper.Run()

			Consistently(spyAgent.Envelopes, 2).Should(HaveLen(0))
		})

		It("should scrape if leadership server returns non 423", func() {
			leadership.setReturnCode(http.StatusInternalServerError)

			scraper = app.NewMetricScraper(cfg, testLogger, spyMetricsClient, emptyNATSConn{})
			go scraper.Run()

			Eventually(func() int {
				return len(spyAgent.Envelopes())
			}).Should(BeNumerically(">", 0))
		})

		It("should scrape if no leadership server endpoint is found", func() {
			cfg.LeadershipServerAddr = ""
			scraper = app.NewMetricScraper(cfg, testLogger, spyMetricsClient, emptyNATSConn{})
			go scraper.Run()

			Eventually(func() int {
				return len(spyAgent.Envelopes())
			}).Should(BeNumerically(">", 0))
		})

		It("should log an error if leadership server fetch returns an error", func() {
			cfg.LeadershipServerAddr = "htp://blablabla.com"
			out := gbytes.NewBuffer()
			scraper = app.NewMetricScraper(cfg, out, spyMetricsClient, emptyNATSConn{})
			go scraper.Run()
			Eventually(out).Should(gbytes.Say("failed to connect to leadership server"))
		})

		It("doesn't not return results if the prom endpoint is slow to respond", func() {
			promServer.setDelay(500 * time.Millisecond)
			cfg.ScrapeTimeout = 250 * time.Millisecond

			scraper = app.NewMetricScraper(cfg, testLogger, spyMetricsClient, emptyNATSConn{})
			go scraper.Run()

			Consistently(func() int {
				return len(spyAgent.Envelopes())
			}, 1).Should(BeNumerically("==", 0))
		})

		It("creates a metric from the number of scrapes", func() {
			promServer.setDelay(500 * time.Millisecond)
			cfg.ScrapeTimeout = 250 * time.Millisecond

			scraper = app.NewMetricScraper(cfg, testLogger, spyMetricsClient, emptyNATSConn{})
			go scraper.Run()

			Eventually(func() bool {
				return spyMetricsClient.HasMetric("num_scrapes", nil)
			}).Should(BeTrue())

			metric := spyMetricsClient.GetMetric("num_scrapes", nil)
			Eventually(metric.Value).Should(BeNumerically(">", 1))
		})

		It("continues attempting scrapes when prom endpoint doesn't exist", func() {
			promServer.stop()

			cfg.ScrapeInterval = 10 * time.Millisecond
			cfg.ScrapeTimeout = 50 * time.Millisecond

			scraper = app.NewMetricScraper(cfg, testLogger, spyMetricsClient, emptyNATSConn{})
			go scraper.Run()

			Eventually(func() bool {
				return spyMetricsClient.HasMetric("num_scrapes", nil)
			}).Should(BeTrue())

			metric := spyMetricsClient.GetMetric("num_scrapes", nil)
			Eventually(metric.Value).Should(BeNumerically(">", 1))
		})

		It("creates a metric for the last total number of attempted scrapes", func() {
			scraper = app.NewMetricScraper(cfg, testLogger, spyMetricsClient, emptyNATSConn{})
			go scraper.Run()

			Eventually(func() bool {
				return spyMetricsClient.HasMetric("last_total_attempted_scrapes", map[string]string{"unit": "total"})
			}).Should(BeTrue())

			metric := spyMetricsClient.GetMetric("last_total_attempted_scrapes", map[string]string{"unit": "total"})
			Eventually(metric.Value).Should(BeNumerically("==", 1))
			Consistently(metric.Value, 1).Should(BeNumerically("==", 1))
		})

		It("creates a metric for the last total number of failed scrapes", func() {
			cfg.ScrapePort = 123456 //Bad port -- scrap fails

			scraper = app.NewMetricScraper(cfg, testLogger, spyMetricsClient, emptyNATSConn{})
			go scraper.Run()

			Eventually(func() bool {
				return spyMetricsClient.HasMetric("last_total_failed_scrapes", map[string]string{"unit": "total"})
			}).Should(BeTrue())

			metric := spyMetricsClient.GetMetric("last_total_failed_scrapes", map[string]string{"unit": "total"})
			Eventually(metric.Value).Should(BeNumerically(">", 0))
			Consistently(metric.Value, 1).Should(BeNumerically("==", 1))
		})

		It("creates a metric for the last total scrape duration", func() {
			scraper = app.NewMetricScraper(cfg, testLogger, spyMetricsClient, emptyNATSConn{})
			go scraper.Run()

			Eventually(func() bool {
				return spyMetricsClient.HasMetric("last_total_scrape_duration", map[string]string{"unit": "ms"})
			}).Should(BeTrue())

			metric := spyMetricsClient.GetMetric("last_total_scrape_duration", map[string]string{"unit": "ms"})
			Eventually(metric.Value).Should(BeNumerically(">", 0))
		})

		It("publishes scrape targets to NATS each scrape", func() {
			spyNATSConn = newSpyNATSConn()

			scraper = app.NewMetricScraper(cfg, testLogger, spyMetricsClient, spyNATSConn)
			go scraper.Run()

			Eventually(spyNATSConn.subj).Should(Receive(Equal("metrics.scrape_targets")))
			Eventually(spyNATSConn.data).Should(Receive(MatchYAML(
				fmt.Sprintf(`{"targets":["%v"], "source":"default-id", "labels":{"deployment": "my-deployment-name", "instance_group": "my-instance-group-name", "id": "default-source", "ip":"%v"}}`,
					fmt.Sprintf("%v/metrics", promServer.url()), promAddr,
				))))
		})
	})
})

type spyNATSConn struct {
	subj chan string
	data chan []byte
}

func (s *spyNATSConn) Publish(subj string, data []byte) error {
	s.subj <- subj
	s.data <- data
	return nil
}

func newSpyNATSConn() *spyNATSConn {
	return &spyNATSConn{
		subj: make(chan string, 100),
		data: make(chan []byte, 100),
	}
}

type emptyNATSConn struct{}

func (emptyNATSConn) Publish(string, []byte) error {
	return nil
}

func buildTags(sourceID, ip string) map[string]string {
	return map[string]string{
		"deployment":     "my-deployment-name",
		"ip":             ip,
		"id":             "default-source",
		"instance_group": "my-instance-group-name",
	}
}

func buildGauge(sourceID, name, ip string, value float64) *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		Tags:     buildTags(sourceID, ip),
		SourceId: sourceID,
		Message: &loggregator_v2.Envelope_Gauge{
			Gauge: &loggregator_v2.Gauge{
				Metrics: map[string]*loggregator_v2.GaugeValue{
					name: {Value: value},
				},
			},
		},
	}
}

func buildCounter(sourceID, name, ip string, value float64) *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		Tags:     buildTags(sourceID, ip),
		SourceId: sourceID,
		Message: &loggregator_v2.Envelope_Counter{
			Counter: &loggregator_v2.Counter{
				Name:  name,
				Total: uint64(value),
			},
		},
	}
}

func createDNSFile(URL string) string {
	contents := fmt.Sprintf(dnsFileTemplate, URL)
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}

	tmpfn := filepath.Join(dir, "records.json")
	tmpfn, err = filepath.Abs(tmpfn)
	Expect(err).ToNot(HaveOccurred())

	//nolint:gosec
	if err := ioutil.WriteFile(tmpfn, []byte(contents), 0666); err != nil {
		log.Fatal(err)
	}
	return tmpfn
}

const (
	promOutput = `
# HELP node_timex_pps_calibration_total Pulse per second count of calibration intervals.
# TYPE node_timex_pps_calibration_total counter
node_timex_pps_calibration_total{source_id="source-1"} 1
# HELP node_timex_pps_error_total Pulse per second count of calibration errors.
# TYPE node_timex_pps_error_total counter
node_timex_pps_error_total{source_id="source-1"} 2
# HELP node_timex_pps_frequency_hertz Pulse per second frequency.
# TYPE node_timex_pps_frequency_hertz gauge
node_timex_pps_frequency_hertz{source_id="source-1"} 3
# HELP node_timex_pps_jitter_seconds Pulse per second jitter.
# TYPE node_timex_pps_jitter_seconds gauge
node_timex_pps_jitter_seconds{source_id="source-2"} 4
# HELP node_timex_pps_jitter_total Pulse per second count of jitter limit exceeded events.
# TYPE node_timex_pps_jitter_total counter
node_timex_pps_jitter_total 5
`
	dnsFileTemplate = `
{
  "record_keys": [
    "id",
    "num_id",
    "instance_group",
    "group_ids",
    "az",
    "az_id",
    "network",
    "network_id",
    "deployment",
    "ip",
    "domain",
    "agent_id",
    "instance_index"
  ],
  "record_infos": [
    [
      "default-source",
      "12345",
      "my-instance-group-name",
      [
        "2345"
      ],
      "my-az-name",
      "34",
      "my-network-name",
      "45",
      "my-deployment-name",
      %q,
      "bosh",
      "6615c4f0-9a52-4ba0-b15c-6534b9bd99a9",
	  1
    ]
  ]
}`
)

type promServer struct {
	sync.Mutex
	delay  time.Duration
	server *httptest.Server
}

func newPromServer() *promServer {
	return &promServer{}
}

func (s *promServer) setDelay(d time.Duration) {
	s.Lock()
	defer s.Unlock()

	s.delay = d
}

func (s *promServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	Expect(r.URL.Path).To(Equal("/metrics"))
	if s.delay > 0 {
		s.Lock()
		toSleep := s.delay
		s.Unlock()
		time.Sleep(toSleep)
	}

	_, err := w.Write([]byte(promOutput))
	Expect(err).NotTo(HaveOccurred())
}

func (s *promServer) start(testCerts *testhelper.TestCerts) {
	s.server = httptest.NewUnstartedServer(http.HandlerFunc(s.handleRequest))

	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithInternalServiceDefaults(),
		tlsconfig.WithIdentityFromFile(
			testCerts.Cert("system-metrics-agent"),
			testCerts.Key("system-metrics-agent"),
		),
	).Server(
		tlsconfig.WithClientAuthenticationFromFile(testCerts.CA()),
	)

	Expect(err).ToNot(HaveOccurred())

	s.server.TLS = tlsConfig

	s.server.StartTLS()
}

func (s *promServer) stop() {
	s.server.Close()
}

func (s *promServer) url() string {
	return s.server.URL
}

type spyAgent struct {
	loggregator_v2.IngressServer

	mu        sync.Mutex
	envelopes []*loggregator_v2.Envelope
	addr      string
}

func newSpyAgent(testCerts *testhelper.TestCerts) *spyAgent {
	agent := &spyAgent{}

	serverCreds, err := newServerCredentials(
		testCerts.Cert("metron"),
		testCerts.Key("metron"),
		testCerts.CA(),
	)
	if err != nil {
		panic(err)
	}

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	agent.addr = lis.Addr().String()

	grpcServer := grpc.NewServer(grpc.Creds(serverCreds))
	loggregator_v2.RegisterIngressServer(grpcServer, agent)

	go grpcServer.Serve(lis) //nolint:errcheck

	return agent
}

func (s *spyAgent) BatchSender(srv loggregator_v2.Ingress_BatchSenderServer) error {
	for {
		batch, err := srv.Recv()
		if err != nil {
			return err
		}

		for _, e := range batch.GetBatch() {
			if e.GetTimestamp() == 0 {
				panic("0 timestamp!?")
			}

			// We want to make our lives easier for matching against envelopes
			e.Timestamp = 0
		}

		s.mu.Lock()
		s.envelopes = append(s.envelopes, batch.GetBatch()...)
		s.mu.Unlock()
	}
}

func (s *spyAgent) Envelopes() []*loggregator_v2.Envelope {
	s.mu.Lock()
	defer s.mu.Unlock()

	results := make([]*loggregator_v2.Envelope, len(s.envelopes))
	copy(results, s.envelopes)
	return results
}

type spyLeadership struct {
	sync.Mutex
	statusCode int
	server     *httptest.Server
}

func (l *spyLeadership) setReturnCode(code int) {
	l.Lock()
	defer l.Unlock()

	l.statusCode = code
}

func (l *spyLeadership) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.Lock()
	defer l.Unlock()

	w.WriteHeader(l.statusCode)
}

func newSpyLeadership(testCerts *testhelper.TestCerts) *spyLeadership {
	leadership := &spyLeadership{
		statusCode: http.StatusOK,
	}

	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithIdentityFromFile(
			testCerts.Cert("leadership_election"),
			testCerts.Key("leadership_election"),
		)).Server(tlsconfig.WithClientAuthenticationFromFile(testCerts.CA()))
	if err != nil {
		panic(err)
	}

	leadership.server = httptest.NewUnstartedServer(leadership)
	leadership.server.TLS = tlsConfig
	leadership.server.StartTLS()

	return leadership
}

func newServerCredentials(
	certFile string,
	keyFile string,
	caCertFile string,
) (credentials.TransportCredentials, error) {
	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithInternalServiceDefaults(),
		tlsconfig.WithIdentityFromFile(certFile, keyFile),
	).Server(
		tlsconfig.WithClientAuthenticationFromFile(caCertFile),
	)

	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(tlsConfig), nil
}
