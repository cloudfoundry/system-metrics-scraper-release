module code.cloudfoundry.org/system-metrics-scraper

go 1.12

require (
	code.cloudfoundry.org/go-diodes v0.0.0-20190809170250-f77fb823c7ee // indirect
	code.cloudfoundry.org/go-envstruct v1.5.0
	code.cloudfoundry.org/go-loggregator v0.0.0-20190813173818-049b6bf8152a // pinned
	code.cloudfoundry.org/go-metric-registry v0.0.0-20200413202920-40d97c8804ec
	code.cloudfoundry.org/tlsconfig v0.0.0-20200131000646-bbe0f8da39b3
	github.com/armon/go-metrics v0.3.3 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.1 // indirect
	github.com/hashicorp/go-hclog v0.13.0
	github.com/hashicorp/go-immutable-radix v1.2.0 // indirect
	github.com/hashicorp/go-msgpack v1.1.5 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/raft v1.1.2
	github.com/kr/pretty v0.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/nats-io/nats-server/v2 v2.1.6 // indirect
	github.com/nats-io/nats.go v1.9.2
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.10.0
	github.com/prometheus/client_golang v1.6.0 // indirect
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.9.1
	golang.org/x/crypto v0.0.0-20200429183012-4b2356b1ed79 // indirect
	golang.org/x/net v0.0.0-20200506145744-7e3656a0809f // indirect
	golang.org/x/sys v0.0.0-20200509044756-6aff5f38e54f // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20200507105951-43844f6eee31 // indirect
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.2.8
)

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190801041406-cbf593c0f2f3
