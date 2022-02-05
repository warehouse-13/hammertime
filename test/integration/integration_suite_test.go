package integration_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		cliBin, err = gexec.Build("github.com/Callisto13/hammertime")
		Expect(err).NotTo(HaveOccurred())

		serverBin, err = gexec.Build("github.com/Callisto13/hammertime/test/fakeserver")
		Expect(err).NotTo(HaveOccurred())

		serverCmd := exec.Command(serverBin)
		serverSession, err = gexec.Start(serverCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		serverSession.Terminate().Wait()
		gexec.CleanupBuildArtifacts()
	})

	RunSpecs(t, "Integration Suite")
}
