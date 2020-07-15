package app

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"gopkg.in/yaml.v2"

	metrics "code.cloudfoundry.org/go-metric-registry"
	"code.cloudfoundry.org/tlsconfig"

	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/system-metrics-scraper/pkg/scraper"
)

type MetricScraper struct {
	cfg           Config
	log           *log.Logger
	scrapeTargets scraper.TargetProvider
	doneChan      chan struct{}
	stoppedChan   chan struct{}
	metrics       metricsClient
	natsConn      natsConn
}

type metricsClient interface {
	NewCounter(name, helpText string, opts ...metrics.MetricOption) metrics.Counter
	NewGauge(name, helpText string, opts ...metrics.MetricOption) metrics.Gauge
}

type natsConn interface {
	Publish(string, []byte) error
}

func NewMetricScraper(cfg Config, w io.Writer, m metricsClient, n natsConn) *MetricScraper {
	return &MetricScraper{
		cfg:           cfg,
		log:           log.New(w, "", log.LstdFlags),
		scrapeTargets: scraper.NewDNSScrapeTargetProvider(cfg.DefaultSourceID, cfg.DNSFile, cfg.ScrapePort),
		doneChan:      make(chan struct{}),
		metrics:       m,
		stoppedChan:   make(chan struct{}),
		natsConn:      n,
	}
}

func (m *MetricScraper) Run() {
	m.scrape()
}

func (m *MetricScraper) scrape() {
	creds, err := loggregator.NewIngressTLSConfig(
		m.cfg.CACertPath,
		m.cfg.ClientCertPath,
		m.cfg.ClientKeyPath,
	)
	if err != nil {
		m.log.Fatal(err)
	}

	client, err := loggregator.NewIngressClient(
		creds,
		loggregator.WithAddr(m.cfg.LoggregatorIngressAddr),
		loggregator.WithLogger(m.log),
	)
	if err != nil {
		m.log.Fatal(err)
	}

	tlsClient := newTLSClient(m.cfg)
	s := scraper.New(
		m.scrapeTargets,
		client,
		func(addr string, _ map[string]string) (response *http.Response, e error) {
			return tlsClient.Get(addr)
		},
		m.cfg.DefaultSourceID,
		scraper.WithMetricsClient(m.metrics),
	)

	leadershipClient := m.leadershipClient()
	numScrapes := m.metrics.NewCounter(
		"num_scrapes",
		"Total number of scrapes performed by the metric scraper.",
	)
	t := time.NewTicker(m.cfg.ScrapeInterval)
	for {
		select {
		case <-t.C:
			resp, err := leadershipClient.Get(m.cfg.LeadershipServerAddr)
			if err != nil {
				m.log.Printf("failed to connect to leadership server: %s\n", err)
			}

			if err == nil && resp.StatusCode == http.StatusLocked {
				continue
			}

			for _, t := range m.scrapeTargets() {
				newT := target{
					Targets: []string{t.MetricURL},
					Labels:  t.DefaultTags,
					Source:  t.ID,
				}

				bytes, err := yaml.Marshal(newT)
				if err != nil {
					m.log.Printf("unable to marshal target(%s): %s\n", t.MetricURL, err)
					continue
				}

				err = m.natsConn.Publish(scrapeTargetQueueName, bytes)
				if err != nil {
					m.log.Printf("failed to publish targets: %s", err)
				}
			}

			if err = s.Scrape(); err != nil {
				m.log.Printf("failed to scrape: %s", err)
			}

			numScrapes.Add(1.0)
		case <-m.doneChan:
			close(m.stoppedChan)
			return
		}
	}
}

const scrapeTargetQueueName = "metrics.scrape_targets"

type target struct {
	Targets []string          `json:"targets",yaml:"targets"`
	Labels  map[string]string `json:"labels",yaml:"labels"`
	Source  string            `json:"-",yaml:"source"`
}

func (m *MetricScraper) leadershipClient() *http.Client {
	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithIdentityFromFile(m.cfg.LeadershipCertPath, m.cfg.LeadershipKeyPath),
	).Client(
		tlsconfig.WithAuthorityFromFile(m.cfg.LeadershipCACertPath),
		tlsconfig.WithServerName("leadership_election"),
	)
	if err != nil {
		m.log.Fatalf("failed to generate leadership election client tls config: %s", err)
	}

	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   5 * time.Second,
	}
}

func (m *MetricScraper) Stop() {
	close(m.doneChan)
	<-m.stoppedChan
}

func newTLSClient(cfg Config) *http.Client {
	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithInternalServiceDefaults(),
		tlsconfig.WithIdentityFromFile(cfg.MetricsCertPath, cfg.MetricsKeyPath),
	).Client(
		tlsconfig.WithAuthorityFromFile(cfg.MetricsCACertPath),
		tlsconfig.WithServerName(cfg.MetricsCN),
	)

	if err != nil {
		log.Panicf("failed to load API client certificates: %s", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   cfg.ScrapeTimeout,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       tlsConfig,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.ScrapeTimeout,
	}
}
