package microvm_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks/flintlock/api/types"
	"github.com/weaveworks/flintlock/client/cloudinit/instance"
	"github.com/weaveworks/flintlock/client/cloudinit/userdata"
	"gopkg.in/yaml.v2"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/client/fakeclient"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/microvm"
)

var _ = Describe("Client", func() {
	var (
		name       = "Pantalaimon"
		namespace  = "Casper"
		mockClient *fakeclient.FakeMicroVMClient
		c          client.Client
		m          microvm.MicroVMManager
	)

	BeforeEach(func() {
		mockClient = new(fakeclient.FakeMicroVMClient)
		c = client.New(mockClient)
		m = microvm.NewManager(c)
	})

	It("creates a MicroVm", func() {
		microVm := &v1alpha1.CreateMicroVMResponse{
			Microvm: &types.MicroVM{
				Spec: &types.MicroVMSpec{
					Id:        name,
					Namespace: namespace,
				},
			},
		}

		cfg := &config.Config{
			MvmName:      name,
			MvmNamespace: namespace,
		}

		mockClient.CreateMicroVMReturns(microVm, nil)
		result, err := m.Create(cfg)
		Expect(err).ToNot(HaveOccurred())

		_, inputReq, _ := mockClient.CreateMicroVMArgsForCall(0)
		Expect(inputReq.Microvm.Id).To(Equal(name))
		Expect(inputReq.Microvm.Namespace).To(Equal(namespace))

		var userData userdata.UserData
		data, err := base64.StdEncoding.DecodeString(inputReq.Microvm.Metadata["user-data"])
		Expect(err).ToNot(HaveOccurred())
		Expect(yaml.Unmarshal(data, &userData)).To(Succeed())
		Expect(userData.HostName).To(Equal(name))
		Expect(userData.Users[0].Name).To(Equal("root"))

		var metaData instance.Metadata
		data, err = base64.StdEncoding.DecodeString(inputReq.Microvm.Metadata["meta-data"])
		Expect(err).ToNot(HaveOccurred())
		Expect(yaml.Unmarshal(data, &metaData)).To(Succeed())
		Expect(metaData["instance_id"]).To(Equal(fmt.Sprintf("%s/%s", namespace, name)))
		Expect(metaData["local_hostname"]).To(Equal(name))
		Expect(metaData["platform"]).To(Equal("liquid_metal"))

		Expect(mockClient.CreateMicroVMCallCount()).To(Equal(1))
		Expect(result.Microvm.Spec.Id).To(Equal(name))
		Expect(result.Microvm.Spec.Namespace).To(Equal(namespace))
	})

	Context("when an sshkey file is set", func() {
		Context("when file exists", func() {
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

			It("creates a MicroVm", func() {
				microVm := &v1alpha1.CreateMicroVMResponse{
					Microvm: &types.MicroVM{
						Spec: &types.MicroVMSpec{
							Id:        name,
							Namespace: namespace,
						},
					},
				}

				cfg := &config.Config{
					MvmName:      name,
					MvmNamespace: namespace,
					SSHKeyPath:   keyFile.Name(),
				}

				mockClient.CreateMicroVMReturns(microVm, nil)
				_, err := m.Create(cfg)
				Expect(err).ToNot(HaveOccurred())

				_, inputReq, _ := mockClient.CreateMicroVMArgsForCall(0)
				Expect(inputReq.Microvm.Id).To(Equal(name))
				Expect(inputReq.Microvm.Namespace).To(Equal(namespace))
				var user userdata.UserData
				userData, err := base64.StdEncoding.DecodeString(inputReq.Microvm.Metadata["user-data"])
				Expect(err).ToNot(HaveOccurred())
				Expect(yaml.Unmarshal(userData, &user)).To(Succeed())
				Expect(user.Users[0].Name).To(Equal("root"))
				Expect(user.Users[0].SSHAuthorizedKeys[0]).To(Equal("this is a test key woohoo"))

				Expect(mockClient.CreateMicroVMCallCount()).To(Equal(1))
			})
		})

		Context("when file does not exist", func() {
			It("returns an error", func() {
				cfg := &config.Config{
					SSHKeyPath: "key.txt",
				}

				_, err := m.Create(cfg)
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
			})
		})
	})

	Context("jsonSpec is set", func() {
		var (
			jsonSpec  = "./../../example.json"
			name      = "mvm1"
			namespace = "ns1"
		)

		It("creates a MicroVm", func() {
			microVm := &v1alpha1.CreateMicroVMResponse{
				Microvm: &types.MicroVM{
					Spec: &types.MicroVMSpec{
						Id:        name,
						Namespace: namespace,
					},
				},
			}

			cfg := &config.Config{
				JSONFile: jsonSpec,
			}

			mockClient.CreateMicroVMReturns(microVm, nil)
			result, err := m.Create(cfg)
			Expect(err).ToNot(HaveOccurred())

			_, inputReq, _ := mockClient.CreateMicroVMArgsForCall(0)
			Expect(inputReq.Microvm.Id).To(Equal(name))
			Expect(inputReq.Microvm.Namespace).To(Equal(namespace))

			Expect(mockClient.CreateMicroVMCallCount()).To(Equal(1))
			Expect(result.Microvm.Spec.Id).To(Equal(name))
			Expect(result.Microvm.Spec.Namespace).To(Equal(namespace))
		})

		Context("when file does not exist", func() {
			It("returns an error", func() {
				cfg := &config.Config{
					JSONFile: "example1.json",
				}

				_, err := m.Create(cfg)
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
			})
		})
	})
})
