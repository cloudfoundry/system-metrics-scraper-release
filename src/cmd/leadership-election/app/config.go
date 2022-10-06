package app

import (
	envstruct "code.cloudfoundry.org/go-envstruct"
	"code.cloudfoundry.org/system-metrics-scraper/pkg/config"
)

type Config struct {
	// Port is the HTTP port that the agent will bind to for localhost (e.g,
	// http://localhost:<port>).
	Port uint16 `env:"PORT"`

	// HealthPort is the port where pprof and expvar will be bound to.
	HealthPort uint16 `env:"HEALTH_PORT"`

	// NodeIndex determines what data the node stores. It splits up the
	// range
	// of 0 - 18446744073709551615 evenly. If data falls out of range
	// of the
	// given node, it will be routed to theh correct one.
	NodeIndex int `env:"NODE_INDEX"`

	// NodeAddrs are all the addresses (including the current address). They
	// are in order according to their NodeIndex.
	//
	// If NodeAddrs is emptpy or size 1, then data is not routed as it is
	// assumed that the current node is the only one.
	NodeAddrs []string `env:"NODE_ADDRS"`

	CAFile   string `env:"CA_FILE, required, report"`
	CertFile string `env:"CERT_FILE, required, report"`
	KeyFile  string `env:"KEY_FILE, required, report"`

	MetricsServer config.MetricsServer
}

func LoadConfig() (Config, error) {
	cfg := Config{
		Port:       8080,
		HealthPort: 6060,
	}

	if err := envstruct.Load(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
