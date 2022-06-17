package app_test

import (
	"bytes"
	"log"
	"os"

	"code.cloudfoundry.org/go-envstruct"
	"code.cloudfoundry.org/system-metrics-scraper/cmd/metric-scraper/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("configuration", func() {

	var requiredVars = []string{
		"CA_CERT_PATH",
		"CLIENT_CERT_PATH",
		"CLIENT_KEY_PATH",
		"DEFAULT_SOURCE_ID",
		"DNS_FILE",
		"LEADERSHIP_ELECTION_CA_CERT_PATH",
		"LEADERSHIP_ELECTION_CERT_PATH",
		"LEADERSHIP_ELECTION_KEY_PATH",
		"LEADERSHIP_SERVER_ADDR",
		"LOGGREGATOR_AGENT_ADDR",
		"METRICS_CA_FILE_PATH",
		"METRICS_CERT_FILE_PATH",
		"METRICS_KEY_FILE_PATH",
		"NATS_CA_PATH",
		"NATS_CERT_PATH",
		"NATS_KEY_PATH",
		"SCRAPE_PORT",
		"SYSTEM_METRICS_CA_CERT_PATH",
		"SYSTEM_METRICS_CA_CN",
		"SYSTEM_METRICS_CERT_PATH",
		"SYSTEM_METRICS_KEY_PATH",
	}

	BeforeEach(func() {
		err := os.Setenv("NATS_HOSTS", "some-secret")
		Expect(err).ToNot(HaveOccurred())

		for _, v := range requiredVars {
			err := os.Setenv(v, "1234")
			Expect(err).ToNot(HaveOccurred())
		}
	})
	AfterEach(func() {
		err := os.Unsetenv("NATS_HOSTS")
		Expect(err).ToNot(HaveOccurred())

		for _, v := range requiredVars {
			err := os.Unsetenv(v)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("does not report the value of the NATS_HOSTS environment variable", func() {
		var output bytes.Buffer
		envstruct.ReportWriter = &output
		logger := log.New(GinkgoWriter, "", log.LstdFlags)
		app.LoadConfig(logger)
		Expect(output.String()).ToNot(ContainSubstring("some-secret"))
	})

})
