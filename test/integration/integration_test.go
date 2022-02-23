package integration_test

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks/flintlock/api/types"
)

var _ = Describe("Integration", func() {
	var (
		name             = "Pantalaimon"
		namespace        = "Casper"
		defaultName      = "mvm0"
		defaultNamespace = "ns0"
		jsonFile         = "./../../example.json"
	)

	AfterEach(func() {
		cmd := command{action: "delete", args: []string{"--all"}}
		Eventually(executeCommand(cmd)).Should(gexec.Exit(0))
	})

	Context("Create", func() {
		It("can create a default microVM", func() {
			cmd := command{action: "create"}
			session := executeCommand(cmd)
			Eventually(session).Should(gexec.Exit(0))

			var result v1alpha1.CreateMicroVMResponse
			Expect(json.Unmarshal(session.Wait().Out.Contents(), &result)).To(Succeed())
			Expect(result.Microvm.Spec.Id).To(Equal(defaultName))
			Expect(result.Microvm.Spec.Namespace).To(Equal(defaultNamespace))
		})

		It("can create a microVM with a specified name and namespace", func() {
			cmd := command{action: "create", args: []string{"--name", name, "--namespace", namespace}}
			session := executeCommand(cmd)
			Eventually(session).Should(gexec.Exit(0))

			var result v1alpha1.CreateMicroVMResponse
			Expect(json.Unmarshal(session.Wait().Out.Contents(), &result)).To(Succeed())
			Expect(result.Microvm.Spec.Id).To(Equal(name))
			Expect(result.Microvm.Spec.Namespace).To(Equal(namespace))
		})

		It("can create a microVM from a file", func() {
			cmd := command{action: "create", args: []string{"--file", jsonFile, "--name", "this will be overriden"}}
			session := executeCommand(cmd)
			Eventually(session).Should(gexec.Exit(0))

			var result v1alpha1.CreateMicroVMResponse
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

				var result v1alpha1.CreateMicroVMResponse
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

		Context("with no args", func() {
			It("gets the MicroVM from the default name and namespace", func() {
				cmd := command{action: "get"}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var get *types.MicroVM
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &get)).To(Succeed())
				Expect(get.Spec.Id).To(Equal(defaultName))
				Expect(get.Spec.Namespace).To(Equal(defaultNamespace))
			})
		})

		Context("when more than one MicroVM in the same name and namespace group exist", func() {
			var result2 v1alpha1.CreateMicroVMResponse

			BeforeEach(func() {
				createCmd := command{action: "create"}
				session := executeCommand(createCmd)
				Eventually(session).Should(gexec.Exit(0))
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &result2)).To(Succeed())
			})

			It("returns the uuids of those MicroVMs", func() {
				cmd := command{action: "get"}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))
				Eventually(session.Out).Should(gbytes.Say("2 MicroVMs found under ns0/mvm0"))
				Eventually(session.Out).Should(gbytes.Say(*result.Microvm.Spec.Uid))
				Eventually(session.Out).Should(gbytes.Say(*result2.Microvm.Spec.Uid))
			})
		})

		Context("when passing id as argument", func() {
			It("gets MicroVm using --id", func() {
				cmd := command{action: "get", args: []string{"--id", *result.Microvm.Spec.Uid}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var getResult v1alpha1.GetMicroVMResponse
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &getResult)).To(Succeed())
				Expect(getResult.Microvm.Spec.Id).To(Equal(result.Microvm.Spec.Id))
			})

			Context("when passing state as argument", func() {
				It("gets the state of the MicroVm using --state", func() {
					cmd := command{action: "get", args: []string{"--id", *result.Microvm.Spec.Uid, "--state"}}
					session := executeCommand(cmd)
					Eventually(session).Should(gexec.Exit(0))
					Eventually(session.Out).Should(gbytes.Say("CREATED"))
				})
			})
		})

		Context("when passing name as argument", func() {
			BeforeEach(func() {
				createCmd := command{action: "create", args: []string{"--name", name}}
				session := executeCommand(createCmd)
				Eventually(session).Should(gexec.Exit(0))
			})

			It("gets MicroVm using --name from default namespace", func() {
				cmd := command{action: "get", args: []string{"--name", name}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var get *types.MicroVM
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &get)).To(Succeed())
				Expect(get.Spec.Id).To(Equal(name))
				Expect(get.Spec.Namespace).To(Equal(defaultNamespace))
			})
		})

		Context("when passing namespace argument", func() {
			BeforeEach(func() {
				createCmd := command{action: "create", args: []string{"--namespace", namespace}}
				session := executeCommand(createCmd)
				Eventually(session).Should(gexec.Exit(0))
			})

			It("gets default named MicroVm from --namespace", func() {
				cmd := command{action: "get", args: []string{"--namespace", namespace}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var get *types.MicroVM
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &get)).To(Succeed())
				Expect(get.Spec.Namespace).To(Equal(namespace))
				Expect(get.Spec.Id).To(Equal(defaultName))
			})
		})

		Context("when passing name and namespace as arguments", func() {
			BeforeEach(func() {
				createCmd := command{action: "create", args: []string{"--namespace", namespace, "--name", name}}
				session := executeCommand(createCmd)
				Eventually(session).Should(gexec.Exit(0))
			})

			It("gets MicroVM using --namespace and --name", func() {
				cmd := command{action: "get", args: []string{"--namespace", namespace, "--name", name}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var get *types.MicroVM
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &get)).To(Succeed())
				Expect(get.Spec.Namespace).To(Equal(namespace))
				Expect(get.Spec.Id).To(Equal(name))
			})
		})

		Context("when passing name and/or namespace and state as arguments", func() {
			BeforeEach(func() {
				createCmd := command{action: "create", args: []string{"--namespace", namespace, "--name", name}}
				session := executeCommand(createCmd)
				Eventually(session).Should(gexec.Exit(0))
			})

			It("returns the state of the microVM", func() {
				cmd := command{action: "get", args: []string{"--namespace", namespace, "--name", name, "--state"}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				Eventually(session.Out).Should(gbytes.Say("CREATED"))
			})
		})

		Context("when passing a json file as argument", func() {
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

			It("gets MicroVm using --file containing UUID", func() {
				cmd := command{action: "get", args: []string{"--id", "this will be ignored", "--file", getFile.Name()}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var get v1alpha1.GetMicroVMResponse
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &get)).To(Succeed())
				Expect(get.Microvm.Spec.Id).To(Equal(result.Microvm.Spec.Id))
			})

			Context("when uuid is not present in the file", func() {
				BeforeEach(func() {
					cmd := command{action: "create", args: []string{"--file", jsonFile, "--name", "this will be overriden"}}
					session := executeCommand(cmd)
					Eventually(session).Should(gexec.Exit(0))
				})

				It("gets MicroVm from the file name/namespace", func() {
					cmd := command{action: "get", args: []string{"--id", "this will be ignored", "--file", jsonFile}}
					session := executeCommand(cmd)
					Eventually(session).Should(gexec.Exit(0))

					var get *types.MicroVM
					Expect(json.Unmarshal(session.Wait().Out.Contents(), &get)).To(Succeed())
					Expect(get.Spec.Id).To(Equal("mvm1"))
					Expect(get.Spec.Namespace).To(Equal("ns1"))
				})
			})

			Context("when name/namespace and uuid are not present", func() {
				BeforeEach(func() {
					dat, err := ioutil.ReadFile(getFile.Name())
					Expect(err).NotTo(HaveOccurred())

					var m *types.MicroVMSpec
					Expect(json.Unmarshal([]byte(dat), &m)).To(Succeed())

					m.Uid = nil
					m.Id = ""
					m.Namespace = ""

					var data []byte
					data, err = json.Marshal(m)
					Expect(err).NotTo(HaveOccurred())

					Expect(ioutil.WriteFile(getFile.Name(), data, 0)).To(Succeed())
				})

				It("prints the error", func() {
					cmd := command{action: "get", args: []string{"--file", getFile.Name()}}
					session := executeCommand(cmd)
					Eventually(session).Should(gexec.Exit(1))
					Eventually(session.Err).Should(gbytes.Say("required: uuid or name/namespace"))
				})
			})
		})
	})

	Context("Delete", func() {
		var (
			result  v1alpha1.CreateMicroVMResponse
			result2 v1alpha1.CreateMicroVMResponse
			name    = "DELETEME"
		)

		BeforeEach(func() {
			createCmd := command{action: "create", args: []string{"--name", name}}
			session := executeCommand(createCmd)
			Eventually(session).Should(gexec.Exit(0))
			Expect(json.Unmarshal(session.Wait().Out.Contents(), &result)).To(Succeed())

			createCmd = command{action: "create", args: []string{"--name", "leave-me-here"}}
			session = executeCommand(createCmd)
			Eventually(session).Should(gexec.Exit(0))
			Expect(json.Unmarshal(session.Wait().Out.Contents(), &result2)).To(Succeed())
		})

		It("deletes MicroVm by --id", func() {
			cmd := command{action: "delete", args: []string{"--id", *result.Microvm.Spec.Uid}}
			Eventually(executeCommand(cmd)).Should(gexec.Exit(0))

			getCmd := command{action: "get", args: []string{"--id", *result.Microvm.Spec.Uid}}
			getSession := executeCommand(getCmd)
			Eventually(getSession).Should(gexec.Exit(1))
			Eventually(getSession.Err).Should(gbytes.Say("OHH WHAT A DISASTER"))

			getCmd = command{action: "get", args: []string{"--id", *result2.Microvm.Spec.Uid}}
			getSession = executeCommand(getCmd)
			Eventually(getSession).Should(gexec.Exit(0))
		})

		It("deletes MicroVm by --name and --namespace", func() {
			cmd := command{action: "delete", args: []string{"--name", name, "--namespace", defaultNamespace}}
			Eventually(executeCommand(cmd)).Should(gexec.Exit(0))

			getCmd := command{action: "get", args: []string{"--id", *result.Microvm.Spec.Uid}}
			getSession := executeCommand(getCmd)
			Eventually(getSession).Should(gexec.Exit(1))
			Eventually(getSession.Err).Should(gbytes.Say("OHH WHAT A DISASTER"))

			getCmd = command{action: "get", args: []string{"--id", *result2.Microvm.Spec.Uid}}
			getSession = executeCommand(getCmd)
			Eventually(getSession).Should(gexec.Exit(0))
		})

		Context("when more than one Microvm is found in a namespace/name group", func() {
			BeforeEach(func() {
				createCmd := command{action: "create", args: []string{"--name", name}}
				session := executeCommand(createCmd)
				Eventually(session).Should(gexec.Exit(0))
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &result)).To(Succeed())
			})

			It("returns an list of the found uids", func() {
				cmd := command{action: "delete", args: []string{"--name", name, "--namespace", defaultNamespace}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))
				Eventually(session.Out).Should(gbytes.Say("2 MicroVMs found under ns0/DELETEME:"))
			})
		})

		Context("when name is given but namespace is not", func() {
			It("returns an error", func() {
				cmd := command{action: "delete", args: []string{"--name", name}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say("required: --namespace"))
			})
		})

		Context("when namespace is given but name is not", func() {
			It("returns an error", func() {
				cmd := command{action: "delete", args: []string{"--namespace", defaultNamespace}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say("required: --name"))
			})
		})

		Context("when passing a json file", func() {
			var deleteFile *os.File

			BeforeEach(func() {
				var err error
				content, err := json.Marshal(result.Microvm.Spec)
				Expect(err).NotTo(HaveOccurred())
				deleteFile, err = ioutil.TempFile("", "tempfile")
				Expect(err).NotTo(HaveOccurred())

				Expect(ioutil.WriteFile(deleteFile.Name(), []byte(content), 0)).To(Succeed())
			})

			AfterEach(func() {
				Expect(os.Remove(deleteFile.Name())).To(Succeed())
			})

			It("deletes MicroVm by UUID", func() {
				cmd := command{action: "delete", args: []string{"--id", "this will be overriden", "--file", deleteFile.Name()}}
				Eventually(executeCommand(cmd)).Should(gexec.Exit(0))

				getCmd := command{action: "get", args: []string{"--id", *result.Microvm.Spec.Uid}}
				getSession := executeCommand(getCmd)
				Eventually(getSession).Should(gexec.Exit(1))
				Eventually(getSession.Err).Should(gbytes.Say("OHH WHAT A DISASTER"))

				getCmd = command{action: "get", args: []string{"--id", *result2.Microvm.Spec.Uid}}
				getSession = executeCommand(getCmd)
				Eventually(getSession).Should(gexec.Exit(0))
			})

			Context("when uuid is not present in the file", func() {
				var name, namespace string

				BeforeEach(func() {
					dat, err := ioutil.ReadFile(deleteFile.Name())
					Expect(err).NotTo(HaveOccurred())

					var m *types.MicroVMSpec
					Expect(json.Unmarshal([]byte(dat), &m)).To(Succeed())
					m.Uid = nil
					name = m.Id
					namespace = m.Namespace

					var data []byte
					data, err = json.Marshal(m)
					Expect(err).NotTo(HaveOccurred())
					Expect(ioutil.WriteFile(deleteFile.Name(), data, 0)).To(Succeed())
				})

				It("deletes MicroVm by the name/namespace in the file", func() {
					cmd := command{action: "delete", args: []string{"--id", "this will be ignored", "--file", deleteFile.Name()}}
					Eventually(executeCommand(cmd)).Should(gexec.Exit(0))

					getCmd := command{action: "get", args: []string{"--file", deleteFile.Name()}}
					getSession := executeCommand(getCmd)
					Eventually(getSession).Should(gexec.Exit(1))
					Eventually(getSession.Err).Should(gbytes.Say(fmt.Sprintf("MicroVM %s/%s not found", namespace, name)))
				})
			})

			Context("when name/namespace and uuid are not present", func() {
				BeforeEach(func() {
					dat, err := ioutil.ReadFile(deleteFile.Name())
					Expect(err).NotTo(HaveOccurred())

					var m *types.MicroVMSpec
					Expect(json.Unmarshal([]byte(dat), &m)).To(Succeed())
					m.Uid = nil
					m.Id = ""
					m.Namespace = ""

					var data []byte
					data, err = json.Marshal(m)
					Expect(err).NotTo(HaveOccurred())
					Expect(ioutil.WriteFile(deleteFile.Name(), data, 0)).To(Succeed())
				})

				It("prints the error", func() {
					cmd := command{action: "delete", args: []string{"--file", deleteFile.Name()}}
					session := executeCommand(cmd)
					Eventually(session).Should(gexec.Exit(1))
					Eventually(session.Err).Should(gbytes.Say("required: uuid or name/namespace"))
				})
			})
		})

		Context("--all", func() {
			BeforeEach(func() {
				// create ns0/mvm0
				createCmd := command{action: "create"}
				Eventually(executeCommand(createCmd)).Should(gexec.Exit(0))

				// create ns0/Pantalaimon
				createCmd = command{action: "create", args: []string{"--name", name}}
				Eventually(executeCommand(createCmd)).Should(gexec.Exit(0))

				// create Casper/Pantalaimon
				createCmd = command{action: "create", args: []string{"--namespace", namespace, "--name", name}}
				Eventually(executeCommand(createCmd)).Should(gexec.Exit(0))

				var list v1alpha1.ListMicroVMsResponse
				listCmd := command{action: "list"}
				session := executeCommand(listCmd)
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &list)).To(Succeed())
				Expect(list.Microvm).To(HaveLen(5))
			})

			It("deletes all existing microvms across all namespaces", func() {
				cmd := command{action: "delete", args: []string{"--all"}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var list v1alpha1.ListMicroVMsResponse
				cmd = command{action: "list"}
				session = executeCommand(cmd)
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &list)).To(Succeed())
				// all should be gone
				Expect(list.Microvm).To(HaveLen(0))
			})

			It("deletes all existing microvms in a specific namespace", func() {
				cmd := command{action: "delete", args: []string{"--all", "--namespace", defaultNamespace}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var list v1alpha1.ListMicroVMsResponse
				cmd = command{action: "list"}
				session = executeCommand(cmd)
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &list)).To(Succeed())
				// Casper/Pantalaimon should remain
				Expect(list.Microvm).To(HaveLen(1))
			})

			It("deletes all existing microvms in a specific namespace/name group", func() {
				cmd := command{action: "delete", args: []string{"--all", "--namespace", defaultNamespace, "--name", defaultName}}
				session := executeCommand(cmd)
				Eventually(session).Should(gexec.Exit(0))

				var list v1alpha1.ListMicroVMsResponse
				cmd = command{action: "list"}
				session = executeCommand(cmd)
				Expect(json.Unmarshal(session.Wait().Out.Contents(), &list)).To(Succeed())
				// Casper/Pantalaimon, ns0/Pantalaimon, ns0/leave-me-here and ns0/DELETEME should remain
				Expect(list.Microvm).To(HaveLen(4))
			})

			Context("when --name is provided but --namespace is not", func() {
				It("prints the error", func() {
					cmd := command{action: "delete", args: []string{"--all", "--name", defaultName}}
					session := executeCommand(cmd)
					Eventually(session).Should(gexec.Exit(1))
					Eventually(session.Err).Should(gbytes.Say("required: --namespace"))
				})
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

		Context("when no microVMs exist", func() {
			It("prints an empty object", func() {
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
