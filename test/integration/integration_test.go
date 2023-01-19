package integration_test

import (
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"
)

const (
	timeout  = "30s"
	interval = "5s"
)

var _ = Describe("Integration", func() {
	var (
		defaultName      = "mvm0"
		defaultNamespace = "ns0"

		createSession *gexec.Session
		created1      v1alpha1.CreateMicroVMResponse
	)

	BeforeEach(func() {
		createSession = create()
		Eventually(createSession, timeout, interval).Should(gexec.Exit(0))
		Expect(json.Unmarshal(createSession.Out.Contents(), &created1)).To(Succeed())
	})

	AfterEach(func() {
		_ = delete("--all")
	})

	It("creating a MicroVM", func() {
		Expect(created1.Microvm.Spec.Id).To(Equal(defaultName))
		Expect(created1.Microvm.Spec.Namespace).To(Equal(defaultNamespace))
	})

	It("getting a MicroVM", func() {
		session := get("--id", *created1.Microvm.Spec.Uid)
		Eventually(session, timeout, interval).Should(gexec.Exit(0))

		var getResult types.MicroVM
		Expect(json.Unmarshal(session.Out.Contents(), &getResult)).To(Succeed())
		Expect(getResult.Spec.Id).To(Equal(created1.Microvm.Spec.Id))
	})

	It("listing MicroVMs", func() {
		session := list("--namespace", defaultNamespace, "--name", defaultName)
		Eventually(session, timeout, interval).Should(gexec.Exit(0))

		var list1 v1alpha1.ListMicroVMsResponse
		Expect(json.Unmarshal(session.Out.Contents(), &list1)).To(Succeed())
		Expect(list1.Microvm).To(HaveLen(1))
	})

	It("deleting a MicroVM", func() {
		Eventually(func(g Gomega) {
			session := delete("--id", *created1.Microvm.Spec.Uid)
			g.Expect(session.Wait()).To(gexec.Exit(0))
		}, timeout, interval).Should(Succeed())

		Eventually(func(g Gomega) {
			session := get("--id", *created1.Microvm.Spec.Uid)
			g.Expect(session.Wait()).To(gexec.Exit(1))
			g.Expect(session.Err).To(gbytes.Say("rpc error"))
		}, timeout, interval).Should(Succeed())

		Expect(listAll()).To(Equal(0))
	})
})

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

func listAll() int {
	session := list()
	Eventually(session, timeout, interval).Should(gexec.Exit(0))

	var list v1alpha1.ListMicroVMsResponse
	Expect(json.Unmarshal(session.Out.Contents(), &list)).To(Succeed())
	return len(list.Microvm)
}
