module code.cloudfoundry.org/system-metrics-scraper

go 1.17

require (
	code.cloudfoundry.org/go-envstruct v1.5.0
	code.cloudfoundry.org/go-loggregator v0.0.0-20190813173818-049b6bf8152a // pinned
	code.cloudfoundry.org/go-metric-registry v0.0.0-20191209165758-93cfd5e30bb0
	code.cloudfoundry.org/tlsconfig v0.0.0-20200108215323-551ec42d1f74
	github.com/hashicorp/raft v1.1.1
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/prometheus/client_model v0.1.0
	github.com/prometheus/common v0.7.0
	google.golang.org/grpc v1.26.0
	gopkg.in/yaml.v2 v2.3.0
)

require (
	code.cloudfoundry.org/go-diodes v0.0.0-20190809170250-f77fb823c7ee // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20180905210152-236a6d29298a // indirect
	github.com/armon/go-metrics v0.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/hashicorp/go-hclog v0.11.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/nxadm/tail v1.4.4 // indirect
	github.com/prometheus/client_golang v1.3.0 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/square/certstrap v1.2.0 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/net v0.0.0-20200520004742-59133d7f0dd7 // indirect
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	google.golang.org/genproto v0.0.0-20200108215221-bd8f9a0ef82f // indirect
	google.golang.org/protobuf v1.23.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190801041406-cbf593c0f2f3
