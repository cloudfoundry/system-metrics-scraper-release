module code.cloudfoundry.org/system-metrics-scraper

go 1.12

require (
	code.cloudfoundry.org/go-diodes v0.0.0-20190809170250-f77fb823c7ee // indirect
	code.cloudfoundry.org/go-envstruct v1.5.0
	code.cloudfoundry.org/go-loggregator v0.0.0-20190813173818-049b6bf8152a // pinned
	code.cloudfoundry.org/go-metric-registry v0.0.0-20191004164645-33b67ef0f7d1
	code.cloudfoundry.org/tlsconfig v0.0.0-20191126220907-6c65973656e3
	github.com/armon/go-metrics v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/hashicorp/go-hclog v0.10.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/raft v1.1.1
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	github.com/prometheus/client_golang v1.2.1 // indirect
	github.com/prometheus/client_model v0.0.0-20191202183732-d1d2010b5bee
	github.com/prometheus/common v0.7.0
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/square/certstrap v1.2.0 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3 // indirect
	golang.org/x/net v0.0.0-20191207000613-e7e4b65ae663 // indirect
	golang.org/x/sys v0.0.0-20191206220618-eeba5f6aabab // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135 // indirect
	google.golang.org/genproto v0.0.0-20191206224255-0243a4be9c8f // indirect
	google.golang.org/grpc v1.25.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.7
)

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190801041406-cbf593c0f2f3
