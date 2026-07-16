package app_test

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"

	"code.cloudfoundry.org/tlsconfig"
	"code.cloudfoundry.org/tlsconfig/certtest"
	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/system-metrics-scraper/cmd/leadership-election/app"
)

var run = 10000

var _ = Describe("Agent", func() {
	var (
		agents     map[string]*app.Agent
		httpClient *http.Client

		caFile         string
		serverCertPair certKeyPair
	)

	BeforeEach(func() {
		run += GinkgoParallelProcess() * 15

		agents = make(map[string]*app.Agent)

		var ca *certtest.Authority
		ca, caFile = generateCA("leadershipCA")
		serverCertPair = generateCertKeyPair(ca, "server")
		clientCertPair := generateCertKeyPair(ca, "client")

		tlsConfig, err := tlsconfig.Build(
			tlsconfig.WithIdentityFromFile(clientCertPair.certFile, clientCertPair.keyFile),
		).Client(
			tlsconfig.WithAuthorityFromFile(caFile),
		)
		Expect(err).ToNot(HaveOccurred())

		httpClient = &http.Client{
			Transport: &http.Transport{TLSClientConfig: tlsConfig},
		}
	})

	AfterEach(func() {
		run += 2 * len(agents)
	})

	It("returns a 200 if it is the leader", func() {
		var nodes []string

		for i := 3; i < 6; i++ {
			nodes = append(nodes, fmt.Sprintf("127.0.0.1:%d", run+i))
		}

		agents = startAgents(nodes, caFile, serverCertPair)

		Eventually(getLeaderStatusFunc(agents, httpClient), 10).Should(haveSingleLeader(agents))
		Consistently(getLeaderStatusFunc(agents, httpClient), 3).Should(haveSingleLeader(agents))
	})

	It("chooses a leader even if nodes are DNS entries", func() {
		var nodes []string

		for i := 3; i < 6; i++ {
			nodes = append(nodes, fmt.Sprintf("localhost:%d", run+i))
		}

		agents = startAgents(nodes, caFile, serverCertPair)

		Eventually(getLeaderStatusFunc(agents, httpClient), 10).Should(haveSingleLeader(agents))
		Consistently(getLeaderStatusFunc(agents, httpClient), 3).Should(haveSingleLeader(agents))
	})

	It("does not produce multiple leaders when a peer becomes unreachable", func() {
		var nodes []string

		for i := 3; i < 6; i++ {
			nodes = append(nodes, fmt.Sprintf("127.0.0.1:%d", run+i))
		}

		agents = startAgents(nodes, caFile, serverCertPair)

		// Establish a stable single leader across all three nodes.
		Eventually(getLeaderStatusFunc(agents, httpClient), 10).Should(haveSingleLeader(agents))

		// Shut down one agent — pick deterministically by sorted address.
		sortedAddrs := make([]string, 0, len(agents))
		for addr := range agents {
			sortedAddrs = append(sortedAddrs, addr)
		}
		sort.Strings(sortedAddrs)

		deadAddr := sortedAddrs[0]
		agents[deadAddr].Shutdown()
		delete(agents, deadAddr)

		// The two remaining agents have quorum (2/3 > 50%). Exactly one must
		// be leader; two concurrent leaders would indicate split-brain.
		Eventually(getLeaderStatusFunc(agents, httpClient), 10).Should(haveSingleLeader(agents))
		Consistently(getLeaderStatusFunc(agents, httpClient), 3).Should(haveSingleLeader(agents))
	})
})

func startAgents(nodes []string, caFile string, serverCertPair certKeyPair) map[string]*app.Agent {
	agents := map[string]*app.Agent{}

	for i := 0; i < 3; i++ {
		a := app.New(
			i,
			nodes,

			// External address
			app.WithPort(run+i),
			app.WithLogger(log.New(GinkgoWriter, fmt.Sprintf("[AGENT %d]", i), log.LstdFlags)),
		)
		a.Start(
			caFile,
			serverCertPair.certFile,
			serverCertPair.keyFile,
		)
		agents[fmt.Sprintf("https://%s/v1/leader", a.Addr())] = a
	}

	return agents
}

func getLeaderStatusFunc(agents map[string]*app.Agent, httpClient *http.Client) func() []int {
	return func() []int {
		var responses []int
		for addr := range agents {
			resp, err := httpClient.Get(addr)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Or(Equal(http.StatusOK), Equal(http.StatusLocked)))

			responses = append(responses, resp.StatusCode)
		}

		return responses
	}
}

func haveSingleLeader(agents map[string]*app.Agent) types.GomegaMatcher {
	var nonLeaders []int
	for i := 1; i < len(agents); i++ {
		nonLeaders = append(nonLeaders, http.StatusLocked)
	}

	return ConsistOf(append(nonLeaders, http.StatusOK))
}

type certKeyPair struct {
	certFile string
	keyFile  string
}

func tmpFile(prefix string, caBytes []byte) string {
	file, err := os.CreateTemp("", prefix)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.Write(caBytes)
	if err != nil {
		log.Fatal(err)
	}

	return file.Name()
}

func generateCA(caName string) (*certtest.Authority, string) {
	ca, err := certtest.BuildCA(caName)
	if err != nil {
		log.Fatal(err)
	}

	caBytes, err := ca.CertificatePEM()
	if err != nil {
		log.Fatal(err)
	}

	fileName := tmpFile(caName+".crt", caBytes)

	return ca, fileName
}

func generateCertKeyPair(ca *certtest.Authority, commonName string) certKeyPair {
	cert, err := ca.BuildSignedCertificate(commonName, certtest.WithDomains(commonName))
	if err != nil {
		log.Fatal(err)
	}

	certBytes, keyBytes, err := cert.CertificatePEMAndPrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	certFile := tmpFile(commonName+".crt", certBytes)
	keyFile := tmpFile(commonName+".key", keyBytes)

	return certKeyPair{
		certFile: certFile,
		keyFile:  keyFile,
	}
}
