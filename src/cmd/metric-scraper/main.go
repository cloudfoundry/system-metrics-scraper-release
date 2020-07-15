package main

import (
	"log"
	"os"

	metrics "code.cloudfoundry.org/go-metric-registry"

	"code.cloudfoundry.org/system-metrics-scraper/cmd/metric-scraper/app"
)

func main() {
	log := log.New(os.Stderr, "", log.LstdFlags)
	log.Printf("starting Metrics Scraper...")
	defer log.Printf("closing Metrics Scraper...")

	cfg := app.LoadConfig(log)

	metricClient := metrics.NewRegistry(
		log,
		metrics.WithTLSServer(
			int(cfg.MetricsServer.Port),
			cfg.MetricsServer.CertFile,
			cfg.MetricsServer.KeyFile,
			cfg.MetricsServer.CAFile,
		),
	)

	app.NewMetricScraper(cfg, os.Stderr, metricClient).Run()
}
