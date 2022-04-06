package integration_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var (
	cliBin  string
	address = "127.0.0.1"
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

		serverCmd := exec.Command(serverBin)
		serverSession, err = gexec.Start(serverCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		// Sometimes the server doesn't start immediately, so we check that an
		// endpoint is reachable before we carry on with the test.
		getCmd := command{action: "get", args: []string{"--id", "foobar"}}
		getSession := executeCommand(getCmd)
		Eventually(getSession, "10s").Should(gexec.Exit(1))
		Eventually(getSession.Err).Should(gbytes.Say("OHH WHAT A DISASTER"))
	})

	AfterSuite(func() {
		serverSession.Terminate().Wait()
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
