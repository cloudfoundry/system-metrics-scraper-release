package scraper_test

import (
	"io/ioutil"

	"code.cloudfoundry.org/system-metrics-scraper/pkg/scraper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DnsIpProvider", func() {
	It("returns metrics urls from the ips returned from the lookup", func() {
		dnsFile := writeScrapeConfig(genericConfig)
		scrapeTargets := scraper.NewDNSScrapeTargetProvider("default-source", dnsFile, 9100)
		target := scrapeTargets()[0]

		Expect(target.ID).To(Equal("default-source"))
		Expect(target.MetricURL).To(Equal("https://10.0.16.27:9100/metrics"))

		defaultTags := target.DefaultTags
		Expect(defaultTags).To(HaveKeyWithValue("ip", "10.0.16.27"))
		Expect(defaultTags).To(HaveKeyWithValue("id", "default-source"))
		Expect(defaultTags).To(HaveKeyWithValue("instance_group", "my-instance-group-name"))
		Expect(defaultTags).To(HaveKeyWithValue("deployment", "my-deployment-name"))
	})
})

func writeScrapeConfig(config string) string {
	f, err := ioutil.TempFile("", "records.json")
	Expect(err).ToNot(HaveOccurred())

	_, err = f.Write([]byte(config))
	Expect(err).ToNot(HaveOccurred())

	return f.Name()
}

var genericConfig = `
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
      "10.0.16.27",
      "bosh",
      "6615c4f0-9a52-4ba0-b15c-6534b9bd99a9",
      0
    ]
  ]
}`
