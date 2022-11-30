package integration_test

import (
	"encoding/json"

	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/warehouse-13/hammertime/pkg/utils"
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
		name             = "foo"
		namespace        = "bar"
		jsonNamespace    = "ns1"
		jsonName         = "mvm1"
		jsonFile         = "./../../example.json"

		// created1 v1alpha1.CreateMicroVMResponse
		// created2 v1alpha1.CreateMicroVMResponse
		// created3 v1alpha1.CreateMicroVMResponse
		// created4 v1alpha1.CreateMicroVMResponse
	)

	AfterEach(func() {
		_ = delete("--all")
	})

	Context("create", func() {
		It("creating a MicroVM with default values", func() {
			session := create()
			Eventually(session, timeout, interval).Should(gexec.Exit(0))

			created := v1alpha1.CreateMicroVMResponse{}
			Expect(json.Unmarshal(session.Out.Contents(), &created)).To(Succeed())
			Expect(created.Microvm.Spec.Id).To(Equal(defaultName))
			Expect(created.Microvm.Spec.Namespace).To(Equal(defaultNamespace))
		})

		It("creating a MicroVM with a set name", func() {
			session := create("--name", name)
			Eventually(session, timeout, interval).Should(gexec.Exit(0))

			created := v1alpha1.CreateMicroVMResponse{}
			Expect(json.Unmarshal(session.Out.Contents(), &created)).To(Succeed())
			Expect(created.Microvm.Spec.Id).To(Equal(name))
			Expect(created.Microvm.Spec.Namespace).To(Equal(defaultNamespace))
		})

		It("creating a MicroVM with a set namespace", func() {
			session := create("--namespace", namespace)
			Eventually(session, timeout, interval).Should(gexec.Exit(0))

			created := v1alpha1.CreateMicroVMResponse{}
			Expect(json.Unmarshal(session.Out.Contents(), &created)).To(Succeed())
			Expect(created.Microvm.Spec.Id).To(Equal(defaultName))
			Expect(created.Microvm.Spec.Namespace).To(Equal(namespace))
		})

		It("creating a MicroVM from a json", func() {
			session := create("--file", jsonFile)
			Eventually(session, timeout, interval).Should(gexec.Exit(0))

			created := v1alpha1.CreateMicroVMResponse{}
			Expect(json.Unmarshal(session.Out.Contents(), &created)).To(Succeed())
			Expect(created.Microvm.Spec.Id).To(Equal(jsonName))
			Expect(created.Microvm.Spec.Namespace).To(Equal(jsonNamespace))
		})
	})

	Context("get", func() {
		It("getting a MicroVM by UUID", func() {
			uuid, err := uuid.NewV4()
			Expect(err).NotTo(HaveOccurred())
			server.Load(newMicrovm(name, namespace, uuid.String()))

			session := get("--id", uuid.String())
			Eventually(session, timeout, interval).Should(gexec.Exit(0))

			var getResult v1alpha1.GetMicroVMResponse
			Expect(json.Unmarshal(session.Out.Contents(), &getResult)).To(Succeed())
			Expect(getResult.Microvm.Spec.Id).To(Equal(name))
		})

		It("getting a MicroVM with a json file", func() {
			uuid, err := uuid.NewV4()
			Expect(err).NotTo(HaveOccurred())
			server.Load(newMicrovm(jsonName, jsonNamespace, uuid.String()))

			session := get("--file", jsonFile)
			Eventually(session, timeout, interval).Should(gexec.Exit(0))

			var getResult types.MicroVM
			Expect(json.Unmarshal(session.Out.Contents(), &getResult)).To(Succeed())
			Expect(getResult.Spec.Id).To(Equal(jsonName))
		})
	})

	Context("list", func() {
		BeforeEach(func() {
			microvms := []*types.MicroVMSpec{
				{
					Id:        defaultName,
					Namespace: defaultNamespace,
				},
				{
					Id:        defaultName,
					Namespace: defaultNamespace,
				},
				{
					Id:        defaultName,
					Namespace: namespace,
				},
			}
			server.Load(microvms...)
		})

		It("listing all MicroVMs in a set namespace/name group", func() {
			session := list("--namespace", namespace, "--name", defaultName)
			Eventually(session, timeout, interval).Should(gexec.Exit(0))

			var list v1alpha1.ListMicroVMsResponse
			Expect(json.Unmarshal(session.Out.Contents(), &list)).To(Succeed())
			Expect(list.Microvm).To(HaveLen(1))
		})

		It("listing all MicroVMs in a namespace group", func() {
			session := list("--namespace", defaultNamespace)
			Eventually(session, timeout, interval).Should(gexec.Exit(0))

			var list v1alpha1.ListMicroVMsResponse
			Expect(json.Unmarshal(session.Out.Contents(), &list)).To(Succeed())
			Expect(list.Microvm).To(HaveLen(2))
		})

		It("listing all MicroVMs across all namespaces", func() {
			session := list()
			Eventually(session, timeout, interval).Should(gexec.Exit(0))

			var list v1alpha1.ListMicroVMsResponse
			Expect(json.Unmarshal(session.Out.Contents(), &list)).To(Succeed())
			Expect(list.Microvm).To(HaveLen(3))
		})
	})

	FContext("delete", func() {
		It("deleting all MicroVMs in a namespace", func() {
			microvms := createALotOfMVMs(defaultNamespace, defaultName)
			microvms = append(microvms, createALotOfMVMs(namespace, name)...)
			server.Load(microvms...)

			Eventually(func(g Gomega) {
				session := delete("--all", "--namespace", defaultNamespace, "--name", defaultName)
				g.Expect(session.Wait()).To(gexec.Exit(0))
			}, timeout, interval).Should(Succeed())

			Expect(listAll()).To(Equal(5))
		})

		It("deleting a MicroVM by UUID", func() {
			microvms := createALotOfMVMs(defaultNamespace, defaultName)
			server.Load(microvms...)

			uid := microvms[0].Uid
			Eventually(func(g Gomega) {
				session := delete("--id", *uid)
				g.Expect(session.Wait()).To(gexec.Exit(0))
			}, timeout, interval).Should(Succeed())

			Eventually(func(g Gomega) {
				session := get("--id", *uid)
				g.Expect(session.Wait()).To(gexec.Exit(1))
				g.Expect(session.Err).To(gbytes.Say("rpc error"))
			}, timeout, interval).Should(Succeed())

			Expect(listAll()).To(Equal(4))
		})
	})

	// By("deleting all MicroVMs in a name/namespace group")
	// Eventually(func(g Gomega) {
	// 	session := delete("--all", "--namespace", namespace, "--name", defaultName)
	// 	g.Expect(session.Wait()).To(gexec.Exit(0))
	// }, timeout, interval).Should(Succeed())

	// Eventually(func(g Gomega) {
	// 	session := get("--id", *created3.Microvm.Spec.Uid)
	// 	g.Expect(session.Wait()).To(gexec.Exit(1))
	// 	g.Expect(session.Err).To(gbytes.Say("rpc error"))
	// }, timeout, interval).Should(Succeed())

	// Expect(listAll()).To(Equal(6))

	// By("deleting a MicroVM with a json file")
	// Eventually(func(g Gomega) {
	// 	session = delete("--file", jsonFile)
	// 	g.Expect(session.Wait()).To(gexec.Exit(0))
	// }, timeout, interval).Should(Succeed())

	// Eventually(func(g Gomega) {
	// 	session = get("--file", jsonFile)
	// 	g.Expect(session.Wait()).To(gexec.Exit(1))
	// 	g.Expect(session.Err).To(gbytes.Say(fmt.Sprintf("MicroVM %s/%s not found", jsonNamespace, jsonName)))
	// }, timeout, interval).Should(Succeed())

	// Expect(listAll()).To(Equal(5))

	// By("deleting all MicroVMs in all namespaces")
	// Eventually(func(g Gomega) {
	// 	session = delete("--all")
	// 	g.Expect(session.Wait()).To(gexec.Exit(0))
	// }, timeout, interval).Should(Succeed())

	// Expect(listAll()).To(Equal(0))
	// })
})

func createALotOfMVMs(namespace, name string) []*types.MicroVMSpec {
	microvms := []*types.MicroVMSpec{}
	for i := 0; i < 5; i++ {
		uid, err := uuid.NewV4()
		if err != nil {
			return nil
		}

		microvms = append(microvms, newMicrovm(namespace, name, uid.String()))
	}

	return microvms
}

func newMicrovm(namespace, name, uid string) *types.MicroVMSpec {
	return &types.MicroVMSpec{
		Id:        name,
		Namespace: namespace,
		Uid:       utils.PointyString(uid),
	}
}

func listAll() int {
	session := list()
	Eventually(session, timeout, interval).Should(gexec.Exit(0))

	var list v1alpha1.ListMicroVMsResponse
	Expect(json.Unmarshal(session.Out.Contents(), &list)).To(Succeed())
	return len(list.Microvm)
}
