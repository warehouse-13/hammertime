package integration_test

import (
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"
)

var _ = Describe("Integration", func() {
	var (
		name             = "Pantalaimon"
		namespace        = "Casper"
		defaultName      = "mvm0"
		defaultNamespace = "ns0"
	)

	Context("Create", func() {
		var result v1alpha1.CreateMicroVMResponse

		AfterEach(func() {
			cmd := command{action: "delete", args: []string{"--id", *result.Microvm.Spec.Uid}}
			Eventually(executeCommand(cmd)).Should(gexec.Exit(0))
		})

		It("can create a default microVM", func() {
			cmd := command{action: "create"}
			session := executeCommand(cmd)
			Eventually(session).Should(gexec.Exit(0))

			Expect(json.Unmarshal(session.Wait().Out.Contents(), &result)).To(Succeed())
			Expect(result.Microvm.Spec.Id).To(Equal(defaultName))
			Expect(result.Microvm.Spec.Namespace).To(Equal(defaultNamespace))
		})

		It("can create a microVM with a specified name and namespace", func() {
			cmd := command{action: "create", args: []string{"--name", name, "--namespace", namespace}}
			session := executeCommand(cmd)
			Eventually(session).Should(gexec.Exit(0))

			Expect(json.Unmarshal(session.Wait().Out.Contents(), &result)).To(Succeed())
			Expect(result.Microvm.Spec.Id).To(Equal(name))
			Expect(result.Microvm.Spec.Namespace).To(Equal(namespace))
		})

		It("can create a microVM from a file", func() {
			cmd := command{action: "create", args: []string{"--file", "./../../example.json", "--name", "this will be overriden"}}
			session := executeCommand(cmd)
			Eventually(session).Should(gexec.Exit(0))

			Expect(json.Unmarshal(session.Wait().Out.Contents(), &result)).To(Succeed())
			Expect(result.Microvm.Spec.Id).To(Equal("mvm1"))
			Expect(result.Microvm.Spec.Namespace).To(Equal("ns1"))
		})

		Context("when passing ssh key file path", func() {
			var (
				keyFile *os.File
				key     = "this is a test key woohoo"
			)

			BeforeEach(func() {
				var err error
				keyFile, err = ioutil.TempFile("", "ssh_key")
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(keyFile.Name(), []byte(key), 0)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				Expect(os.Remove(keyFile.Name())).To(Succeed())
			})

			It("can create a microVM with a ssh key", func() {
				cmd := command{action: "create", args: []string{"--public-key-path", keyFile.Name()}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				Expect(json.Unmarshal(session.Wait().Out.Contents(), &result)).To(Succeed())
				userData, err := b64.StdEncoding.DecodeString(result.Microvm.Spec.Metadata["user-data"])
				Expect(err).NotTo(HaveOccurred())
				Expect(string(userData)).To(ContainSubstring(key))
			})
		})

	})

	Context("Get", func() {
		var result v1alpha1.CreateMicroVMResponse

		BeforeEach(func() {
			createCmd := command{action: "create"}
			session := executeCommand(createCmd)
			Eventually(session).Should(gexec.Exit(0))
			Expect(json.Unmarshal(session.Wait().Out.Contents(), &result)).To(Succeed())
		})

		AfterEach(func() {
			cmd := command{action: "delete", args: []string{"--id", *result.Microvm.Spec.Uid}}
			Eventually(executeCommand(cmd)).Should(gexec.Exit(0))
		})

		It("gets MicroVm", func() {
			cmd := command{action: "get", args: []string{"--id", *result.Microvm.Spec.Uid}}
			session := executeCommand(cmd)
			Eventually(session).Should(gexec.Exit(0))

			var getResult v1alpha1.GetMicroVMResponse

			Expect(json.Unmarshal(session.Wait().Out.Contents(), &getResult)).To(Succeed())
			Expect(getResult.Microvm.Spec.Id).To(Equal(result.Microvm.Spec.Id))
		})

		It("gets the state of the MicroVm", func() {
			cmd := command{action: "get", args: []string{"--id", *result.Microvm.Spec.Uid, "--state"}}
			session := executeCommand(cmd)
			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out, "30s").Should(gbytes.Say("CREATED"))
		})

		Context("when passing a json file", func() {
			var getFile *os.File

			BeforeEach(func() {
				var err error
				content, err := json.Marshal(result.Microvm.Spec)
				Expect(err).NotTo(HaveOccurred())
				getFile, err = ioutil.TempFile("", "tempfile")
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(getFile.Name(), []byte(content), 0)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				Expect(os.Remove(getFile.Name())).To(Succeed())
			})

			It("gets MicroVm", func() {
				cmd := command{action: "get", args: []string{"--id", "this will be overriden", "--file", getFile.Name()}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var getResult v1alpha1.GetMicroVMResponse

				Expect(json.Unmarshal(session.Wait().Out.Contents(), &getResult)).To(Succeed())
				Expect(getResult.Microvm.Spec.Id).To(Equal(defaultName))
			})
		})
	})

	Context("Delete", func() {
		var result v1alpha1.CreateMicroVMResponse

		BeforeEach(func() {
			createCmd := command{action: "create", args: []string{"--name", "DELETEME"}}
			session := executeCommand(createCmd)
			Eventually(session).Should(gexec.Exit(0))
			Expect(json.Unmarshal(session.Wait().Out.Contents(), &result)).To(Succeed())
		})

		It("deletes MicroVm", func() {
			cmd := command{action: "delete", args: []string{"--id", *result.Microvm.Spec.Uid}}
			Eventually(executeCommand(cmd)).Should(gexec.Exit(0))

			getCmd := command{action: "get", args: []string{"--id", *result.Microvm.Spec.Uid}}
			getSession := executeCommand(getCmd)
			Eventually(getSession).Should(gexec.Exit(1))
			Eventually(getSession.Err).Should(gbytes.Say("OHH WHAT A DISASTER"))
		})

		Context("when passing a json file", func() {
			var deleteFile *os.File

			BeforeEach(func() {
				var err error
				content, err := json.Marshal(result.Microvm.Spec)
				Expect(err).NotTo(HaveOccurred())
				deleteFile, err = ioutil.TempFile("", "tempfile")
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(deleteFile.Name(), []byte(content), 0)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				Expect(os.Remove(deleteFile.Name())).To(Succeed())
			})

			It("deletes MicroVm", func() {
				cmd := command{action: "delete", args: []string{"--id", "this will be overriden", "--file", deleteFile.Name()}}
				Eventually(executeCommand(cmd)).Should(gexec.Exit(0))

				getCmd := command{action: "get", args: []string{"--id", *result.Microvm.Spec.Uid}}
				getSession := executeCommand(getCmd)
				Eventually(getSession).Should(gexec.Exit(1))
				Eventually(getSession.Err).Should(gbytes.Say("OHH WHAT A DISASTER"))
			})
		})
	})

	Context("List", func() {
		Context("when microVMs exist", func() {
			var result1 v1alpha1.CreateMicroVMResponse

			BeforeEach(func() {
				createCmd := command{action: "create"}
				session := executeCommand(createCmd)
				Eventually(session).Should(gexec.Exit(0))
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &result1)).To(Succeed())
			})

			AfterEach(func() {
				cmd := command{action: "delete", args: []string{"--id", *result1.Microvm.Spec.Uid}}
				deleteSession := executeCommand(cmd)
				Eventually(deleteSession).Should(gexec.Exit(0))
			})

			It("lists MicroVm", func() {
				cmd := command{action: "list"}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var list v1alpha1.ListMicroVMsResponse
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &list)).To(Succeed())
				Expect(list.Microvm[0].Spec.Uid).To(Equal(result1.Microvm.Spec.Uid))
			})

			Context("when passing namespace as args", func() {
				var result2 v1alpha1.CreateMicroVMResponse

				BeforeEach(func() {
					createCmd := command{action: "create", args: []string{"--namespace", namespace}}
					session := executeCommand(createCmd)
					Eventually(session).Should(gexec.Exit(0))
					Expect(json.Unmarshal(session.Wait().Out.Contents(), &result2)).To(Succeed())
				})

				AfterEach(func() {
					cmd := command{action: "delete", args: []string{"--id", *result2.Microvm.Spec.Uid}}
					Eventually(executeCommand(cmd)).Should(gexec.Exit(0))
				})

				It("lists MicroVm", func() {
					cmd := command{action: "list", args: []string{"--namespace", namespace}}
					session := executeCommand(cmd)
					Eventually(session).Should(gexec.Exit(0))

					var list v1alpha1.ListMicroVMsResponse
					Expect(json.Unmarshal(session.Wait().Out.Contents(), &list)).To(Succeed())
					Expect(list.Microvm[0].Spec.Uid).To(Equal(result2.Microvm.Spec.Uid))
				})

			})
		})

		Context("when no microVM exist", func() {
			It("lists MicroVm", func() {
				cmd := command{action: "list"}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var list v1alpha1.ListMicroVMsResponse
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &list)).To(Succeed())
				Expect(list.Microvm).To(HaveLen(0))
			})
		})
	})
})
