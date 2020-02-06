module code.cloudfoundry.org/system-metrics-scraper

go 1.12

require (
	code.cloudfoundry.org/go-diodes v0.0.0-20190809170250-f77fb823c7ee // indirect
	code.cloudfoundry.org/go-envstruct v1.5.0
	code.cloudfoundry.org/go-loggregator v0.0.0-20190813173818-049b6bf8152a // pinned
	code.cloudfoundry.org/go-metric-registry v0.0.0-20191209165758-93cfd5e30bb0
	code.cloudfoundry.org/tlsconfig v0.0.0-20200131000646-bbe0f8da39b3
	github.com/armon/go-metrics v0.3.0 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/hashicorp/go-hclog v0.12.0
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/raft v1.1.2
	github.com/kr/pretty v0.2.0 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/prometheus/client_golang v1.4.0 // indirect
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.9.1
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5 // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20200205142000-a86caf926a67 // indirect
	google.golang.org/grpc v1.27.1
	gopkg.in/yaml.v2 v2.2.8
)

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190801041406-cbf593c0f2f3
