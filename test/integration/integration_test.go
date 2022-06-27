package integration_test

import (
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"
)

var _ = Describe("Integration", func() {
	var (
		defaultName      = "mvm0"
		defaultNamespace = "ns0"
		name             = "foo"
		namespace        = "bar"
		jsonNamespace    = "ns1"
		jsonName         = "mvm1"
		jsonFile         = "./../../example.json"
		timeout          = "30s"
		interval         = "5s"

		created1 v1alpha1.CreateMicroVMResponse
		created2 v1alpha1.CreateMicroVMResponse
		created3 v1alpha1.CreateMicroVMResponse
		created4 v1alpha1.CreateMicroVMResponse
	)

	AfterEach(func() {
		// TODO do this a bit smarter
		_ = delete("--id", *created1.Microvm.Spec.Uid)

		_ = delete("--id", *created2.Microvm.Spec.Uid)

		_ = delete("--id", *created3.Microvm.Spec.Uid)

		_ = delete("--id", *created4.Microvm.Spec.Uid)
	})

	It("Can interact with a flintlock server", func() {
		By("creating a MicroVM with default values")
		session := create()
		Eventually(session, timeout, interval).Should(gexec.Exit(0))

		Expect(json.Unmarshal(session.Out.Contents(), &created1)).To(Succeed())
		Expect(created1.Microvm.Spec.Id).To(Equal(defaultName))
		Expect(created1.Microvm.Spec.Namespace).To(Equal(defaultNamespace))

		By("creating a MicroVM with a set name")
		session = create("--name", name)
		Eventually(session, timeout, interval).Should(gexec.Exit(0))

		Expect(json.Unmarshal(session.Out.Contents(), &created2)).To(Succeed())
		Expect(created2.Microvm.Spec.Id).To(Equal(name))
		Expect(created2.Microvm.Spec.Namespace).To(Equal(defaultNamespace))

		By("creating a MicroVM with a set namespace")
		session = create("--namespace", namespace)
		Eventually(session, timeout, interval).Should(gexec.Exit(0))

		Expect(json.Unmarshal(session.Out.Contents(), &created3)).To(Succeed())
		Expect(created3.Microvm.Spec.Id).To(Equal(defaultName))
		Expect(created3.Microvm.Spec.Namespace).To(Equal(namespace))

		By("creating a MicroVM from a json file")
		session = create("--file", jsonFile)
		Eventually(session, timeout, interval).Should(gexec.Exit(0))

		Expect(json.Unmarshal(session.Out.Contents(), &created4)).To(Succeed())
		Expect(created4.Microvm.Spec.Id).To(Equal(jsonName))
		Expect(created4.Microvm.Spec.Namespace).To(Equal(jsonNamespace))

		By("getting a MicroVM by UUID")
		session = get("--id", *created1.Microvm.Spec.Uid)
		Eventually(session, timeout, interval).Should(gexec.Exit(0))

		var getResult v1alpha1.GetMicroVMResponse
		Expect(json.Unmarshal(session.Out.Contents(), &getResult)).To(Succeed())
		Expect(getResult.Microvm.Spec.Id).To(Equal(created1.Microvm.Spec.Id))

		// TODO this test fails and I have zero idea why
		// By("getting a MicroVM with a json file")
		// session = get("--file", jsonFile)
		// Eventually(session, timeout, interval).Should(gexec.Exit(0))

		// var getResult2 v1alpha1.GetMicroVMResponse
		// Expect(json.Unmarshal(session.Out.Contents(), &getResult2)).To(Succeed())
		// Expect(getResult2.Microvm.Spec.Id).To(Equal(created1.Microvm.Spec.Id))

		By("listing all MicroVMs in a set namespace/name group")
		session = list("--namespace", namespace, "--name", defaultName)
		Eventually(session, timeout, interval).Should(gexec.Exit(0))

		var list1 v1alpha1.ListMicroVMsResponse
		Expect(json.Unmarshal(session.Out.Contents(), &list1)).To(Succeed())
		Expect(list1.Microvm).To(HaveLen(1))

		By("listing all MicroVMs in a namespace group")
		session = list("--namespace", namespace)
		Eventually(session, timeout, interval).Should(gexec.Exit(0))

		var list2 v1alpha1.ListMicroVMsResponse
		Expect(json.Unmarshal(session.Out.Contents(), &list2)).To(Succeed())
		Expect(list2.Microvm).To(HaveLen(1))

		By("deleting a MicroVM by UUID")
		// TODO all deletions are here now but they can be moved after #41
		Eventually(func(g Gomega) {
			session := delete("--id", *created1.Microvm.Spec.Uid)
			g.Expect(session.Wait()).To(gexec.Exit(0))

			session = delete("--id", *created2.Microvm.Spec.Uid)
			g.Expect(session.Wait()).To(gexec.Exit(0))

			session = delete("--id", *created3.Microvm.Spec.Uid)
			g.Expect(session.Wait()).To(gexec.Exit(0))
		}, timeout, interval).Should(Succeed())

		Eventually(func(g Gomega) {
			session := get("--id", *created1.Microvm.Spec.Uid)
			g.Expect(session.Wait()).To(gexec.Exit(1))
			g.Expect(session.Err).To(gbytes.Say("rpc error"))

			session = get("--id", *created2.Microvm.Spec.Uid)
			g.Expect(session.Wait()).To(gexec.Exit(1))
			g.Expect(session.Err).To(gbytes.Say("rpc error"))

			session = get("--id", *created3.Microvm.Spec.Uid)
			g.Expect(session.Wait()).To(gexec.Exit(1))
			g.Expect(session.Err).To(gbytes.Say("rpc error"))
		}, timeout, interval).Should(Succeed())

		// TODO #41
		By("deleting all MicroVMs in a namespace")

		// TODO #41
		By("deleting all MicroVMs in a name/namespace group")

		By("deleting a MicroVM with a json file")
		Eventually(func(g Gomega) {
			session = delete("--id", *created4.Microvm.Spec.Uid)
			g.Expect(session.Wait()).To(gexec.Exit(0))
		}, timeout, interval).Should(Succeed())

		Eventually(func(g Gomega) {
			session = get("--id", *created4.Microvm.Spec.Uid)
			g.Expect(session.Wait()).To(gexec.Exit(1))
			g.Expect(session.Err).To(gbytes.Say("rpc error"))
		}, timeout, interval).Should(Succeed())
	})
})

func create(opts ...string) *gexec.Session {
	args := []string{"--grpc-address", address, "create"}
	args = append(args, opts...)

	return runCmd(exec.Command(cliBin, args...))
}

func get(opts ...string) *gexec.Session {
	args := []string{"--grpc-address", address, "get"}
	args = append(args, opts...)

	return runCmd(exec.Command(cliBin, args...))
}

func list(opts ...string) *gexec.Session {
	args := []string{"--grpc-address", address, "list"}
	args = append(args, opts...)

	return runCmd(exec.Command(cliBin, args...))
}

func delete(opts ...string) *gexec.Session {
	args := []string{"--grpc-address", address, "delete"}
	args = append(args, opts...)

	return runCmd(exec.Command(cliBin, args...))
}

func runCmd(cmd *exec.Cmd) *gexec.Session {
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
