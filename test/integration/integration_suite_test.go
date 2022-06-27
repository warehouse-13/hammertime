package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var (
	cliBin  string
	address string
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	var (
		serverBin     string
		serverSession *gexec.Session
	)

	BeforeSuite(func() {
		var err error
		cliBin, err = gexec.Build("github.com/warehouse-13/hammertime")
		Expect(err).NotTo(HaveOccurred())

		serverBin, err = gexec.Build("github.com/warehouse-13/hammertime/test/fakeserver")
		Expect(err).NotTo(HaveOccurred())

		if remote_test_server := os.Getenv("TEST_SERVER"); remote_test_server != "" {
			address = remote_test_server
			fmt.Fprintf(GinkgoWriter, "Using real Flintlock server at %s: tests may take a little longer", address)
		} else {
			serverCmd := exec.Command(serverBin)
			serverSession, err = gexec.Start(serverCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			address = "127.0.0.1:9090"
		}

		// Sometimes the server doesn't start immediately, so we check that an
		// endpoint is reachable before we carry on with the test.
		getCmd := command{action: "get", args: []string{"--id", "foobar"}}
		getSession := executeCommand(getCmd)
		Eventually(getSession, "10s").Should(gexec.Exit(1))
		Eventually(getSession.Err).Should(gbytes.Say("rpc error"))
	})

	AfterSuite(func() {
		if serverSession != nil {
			serverSession.Terminate().Wait()
		}

		gexec.CleanupBuildArtifacts()
	})

	RunSpecs(t, "Integration Suite")
}

type command struct {
	action string
	args   []string
}

func executeCommand(command command) *gexec.Session {
	var args = []string{"--grpc-address", address, command.action}
	cmd := exec.Command(cliBin, append(args, command.args...)...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
