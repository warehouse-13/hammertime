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
	"github.com/warehouse-13/safety"
)

var (
	cliBin  string
	address string
	token   = "secret"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	var fakeserver *safety.FakeServer

	BeforeSuite(func() {
		var err error
		cliBin, err = gexec.Build("github.com/warehouse-13/hammertime")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv("AUTH_TOKEN", token)).To(Succeed())

		if remote_test_server := os.Getenv("TEST_SERVER"); remote_test_server != "" {
			address = remote_test_server
			fmt.Fprintf(GinkgoWriter, "Using real Flintlock server at %s: tests may take a little longer", address)
		} else {
			fakeserver = safety.New()
			address = fakeserver.Start("")
		}

		// Sometimes the server doesn't start immediately, so we check that an
		// endpoint is reachable before we carry on with the test.
		getCmd := command{action: "get", args: []string{"--id", "foobar"}}
		getSession := executeCommand(getCmd)
		Eventually(getSession, "10s").Should(gexec.Exit(1))
		Eventually(getSession.Err).Should(gbytes.Say("rpc error"))
	})

	AfterSuite(func() {
		if fakeserver != nil {
			Expect(fakeserver.Stop()).To(Succeed())
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
	var args = []string{command.action, "--grpc-address", address, "--token", token}
	cmd := exec.Command(cliBin, append(args, command.args...)...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
