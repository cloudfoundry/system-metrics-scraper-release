package app

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	metrics "code.cloudfoundry.org/go-metric-registry"
	"code.cloudfoundry.org/tlsconfig"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
)

// Agent is a Leadership Election Agent. It determines if the local process
// should act as a leader or not.
type Agent struct {
	log  *log.Logger
	port int
	lis  net.Listener
	srv  *http.Server
	m    Metrics

	nodeIndex int
	nodes     []string

	r *raft.Raft
}

// New returns a new Agent.
func New(nodeIndex int, nodes []string, opts ...AgentOption) *Agent {
	a := &Agent{
		log:  log.New(io.Discard, "", 0),
		port: 8080,

		nodeIndex: nodeIndex,
		nodes:     nodes,
		m:         NopMetrics{},
	}

	for _, o := range opts {
		o(a)
	}

	return a
}

// AgentOption configures an Agent by overriding defaults.
type AgentOption func(*Agent)

// WithLogger returns an AgentOption that configures the logger for the Agent.
// It defaults to a silent logger.
func WithLogger(log *log.Logger) AgentOption {
	return func(a *Agent) {
		a.log = log
	}
}

// WithPort configures the port to bind the HTTP server to. It will always
// bind to localhost. Defaults to 8080.
func WithPort(port int) AgentOption {
	return func(a *Agent) {
		a.port = port
	}
}

// Metrics registers Gauge metrics.
type Metrics interface {
	NewGauge(name, helpText string, opts ...metrics.MetricOption) metrics.Gauge
}

type NopGauge struct{}

func (n NopGauge) Add(float64) {}

func (n NopGauge) Set(float64) {}

// NopMetrics implements Metrics, but simply discards them.
type NopMetrics struct{}

// NewGauge implements Metrics.
func (m NopMetrics) NewGauge(name, helpText string, opts ...metrics.MetricOption) metrics.Gauge {
	return NopGauge{}
}

// WithMetrics configures the metrics for Agent. Defaults to NopMetrics.
func WithMetrics(m Metrics) AgentOption {
	return func(a *Agent) {
		a.m = m
	}
}

// Start starts the Agent. It does not block.
func (a *Agent) Start(caFile, certFile, keyFile string) {
	tlsConfig := buildTLSConfig(caFile, certFile, keyFile)

	lis, err := tls.Listen("tcp", fmt.Sprintf("localhost:%d", a.port), tlsConfig)
	if err != nil {
		a.log.Fatalf("failed to listen on localhost:%d", a.port)
	}
	a.lis = lis

	setLeadershipStatus := a.m.NewGauge("leadership_status", "1 if this instance is the leader, 0 otherwise.")

	isLeader := a.startRaft()

	go func() {
		for range time.Tick(time.Second) {
			if isLeader() {
				setLeadershipStatus.Set(1)
				continue
			}
			setLeadershipStatus.Set(0)
		}
	}()

	a.srv = leaderStatusServer(isLeader)

	go func() {
		if err := a.srv.Serve(lis); err != nil && err != http.ErrServerClosed {
			a.log.Fatal(err)
		}
	}()
}

// Shutdown stops the Raft cluster and the HTTP server.
func (a *Agent) Shutdown() {
	if a.r != nil {
		_ = a.r.Shutdown()
	}
	if a.srv != nil {
		_ = a.srv.Close()
	}
}

func leaderStatusServer(isLeader func() bool) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/leader", func(w http.ResponseWriter, r *http.Request) {
		if isLeader() {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusLocked)
	})
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
	}
	return srv
}

func buildTLSConfig(caFile, certFile, keyFile string) *tls.Config {
	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithInternalServiceDefaults(),
		tlsconfig.WithIdentityFromFile(certFile, keyFile),
	).Server(
		tlsconfig.WithClientAuthenticationFromFile(caFile),
	)
	if err != nil {
		log.Fatal(err)
	}
	return tlsConfig
}

func (a *Agent) startRaft() func() bool {
	localAddr := a.nodes[a.nodeIndex]
	addr, err := net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		a.log.Fatalf("failed to resolve address %s: %s", localAddr, err)
	}

	transport, err := raft.NewTCPTransportWithLogger(
		localAddr,
		addr,
		100,
		30*time.Second,
		hclog.Default(),
	)
	if err != nil {
		a.log.Fatalf("failed to create raft TCP transport: %s", err)
	}

	a.initRaft(localAddr, transport)

	return func() bool {
		return a.r.Leader() == raft.ServerAddress(addr.String())
	}
}

// initRaft initializes the Raft cluster once at startup. It is never called
// again after Start() returns. The cluster is not rebuilt if peers become
// temporarily unreachable; standard Raft handles unreachable peers internally.
func (a *Agent) initRaft(localAddr string, transport raft.Transport) {
	var peers []raft.Server
	for _, addr := range a.nodes {
		peers = append(peers, raft.Server{
			ID:      raft.ServerID(addr),
			Address: raft.ServerAddress(addr),
		})
	}

	store := raft.NewInmemStore()
	var err error
	a.r, err = raft.NewRaft(
		&raft.Config{
			ProtocolVersion:    raft.ProtocolVersionMax,
			LocalID:            raft.ServerID(localAddr),
			HeartbeatTimeout:   100 * time.Millisecond,
			ElectionTimeout:    1 * time.Second,
			CommitTimeout:      1 * time.Second,
			MaxAppendEntries:   100,
			SnapshotInterval:   time.Second,
			LeaderLeaseTimeout: 100 * time.Millisecond,
			LogOutput:          io.Discard,
		},
		nil,
		store,
		store,
		raft.NewInmemSnapshotStore(),
		transport,
	)
	if err != nil {
		a.log.Fatalf("failed to create raft cluster: %s", err)
	}

	a.r.BootstrapCluster(raft.Configuration{Servers: peers})
}

// Addr returns the address the Agent is listening to for HTTP requests (e.g.,
// 127.0.0.1:8080). It is only valid after calling Start().
func (a *Agent) Addr() string {
	return a.lis.Addr().String()
}
