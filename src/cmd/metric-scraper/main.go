package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"os"
	"time"

	metrics "code.cloudfoundry.org/go-metric-registry"
	"github.com/nats-io/nats.go"

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

	natsConn := connectToNATS(cfg, log)

	app.NewMetricScraper(cfg, log, metricClient, natsConn).Run()
}

func connectToNATS(cfg app.Config, logger *log.Logger) *nats.Conn {
	opts := nats.Options{
		Servers:           cfg.NatsHosts,
		PingInterval:      20 * time.Second,
		AllowReconnect:    true,
		MaxReconnect:      -1,
		ReconnectWait:     100 * time.Millisecond,
		ClosedCB:          closedCB(logger),
		DisconnectedErrCB: disconnectErrHandler(logger),
		ReconnectedCB:     reconnectedCB(logger),
		TLSConfig:         getTLSConfig(cfg),
	}

	natsConn, err := opts.Connect()
	if err != nil {
		logger.Fatalf("Unable to connect to nats servers: %s", err)
	}
	return natsConn
}

func getTLSConfig(cfg app.Config) *tls.Config {
	caCert, err := ioutil.ReadFile(cfg.NatsCAPath)
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("Failed to load CA certificate from file %s", cfg.NatsCAPath)
	}

	cert, err := tls.LoadX509KeyPair(cfg.NatsCertPath, cfg.NatsKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
}

func closedCB(logger *log.Logger) func(conn *nats.Conn) {
	return func(conn *nats.Conn) {
		logger.Println("Nats Connection Closed")
	}
}

func reconnectedCB(logger *log.Logger) func(conn *nats.Conn) {
	return func(conn *nats.Conn) {
		logger.Printf("Reconnected to %s\n", conn.ConnectedUrl())
	}
}

func disconnectErrHandler(logger *log.Logger) func(conn *nats.Conn, err error) {
	return func(conn *nats.Conn, err error) {
		logger.Printf("Nats Error %s\n", err)
	}
}
