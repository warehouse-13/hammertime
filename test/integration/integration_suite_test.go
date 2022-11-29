package integration_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/warehouse-13/hammertime/test/fakeserver"
)

var (
	cliBin  string
	address string
	token   = "secret"

	server = fakeserver.New()
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)

	BeforeSuite(func() {
		var err error
		cliBin, err = gexec.Build("github.com/warehouse-13/hammertime")
		Expect(err).NotTo(HaveOccurred())

		address = server.Start(token)

		// Sometimes the server doesn't start immediately, so we check that an
		// endpoint is reachable before we carry on with the test.
		getSession := get("--id", "foobar")
		Eventually(getSession, "10s").Should(gexec.Exit(1))
		Eventually(getSession.Err).Should(gbytes.Say("rpc error"))
	})

	AfterSuite(func() {
		Expect(server.Stop()).To(Succeed())
		gexec.CleanupBuildArtifacts()
	})

	RunSpecs(t, "Integration Suite")
}

func create(opts ...string) *gexec.Session {
	args := []string{"create", "--grpc-address", address, "--token", token}
	args = append(args, opts...)

	return runCmd(exec.Command(cliBin, args...))
}

func get(opts ...string) *gexec.Session {
	args := []string{"get", "--grpc-address", address, "--token", token}
	args = append(args, opts...)

	return runCmd(exec.Command(cliBin, args...))
}

func list(opts ...string) *gexec.Session {
	args := []string{"list", "--grpc-address", address, "--token", token}
	args = append(args, opts...)

	return runCmd(exec.Command(cliBin, args...))
}

func delete(opts ...string) *gexec.Session {
	args := []string{"delete", "--grpc-address", address, "--token", token}
	args = append(args, opts...)

	return runCmd(exec.Command(cliBin, args...))
}

func runCmd(cmd *exec.Cmd) *gexec.Session {
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
