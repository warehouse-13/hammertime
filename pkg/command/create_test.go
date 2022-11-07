package command_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"
	"github.com/weaveworks-liquidmetal/flintlock/client/cloudinit/instance"
	"github.com/weaveworks-liquidmetal/flintlock/client/cloudinit/userdata"
	"gopkg.in/yaml.v2"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/client/fakeclient"
	"github.com/warehouse-13/hammertime/pkg/command"
	"github.com/warehouse-13/hammertime/pkg/config"
)

var _ = Describe("CreateFn", func() {
	var (
		name       = "Pantalaimon"
		namespace  = "Casper"
		mockClient *fakeclient.FakeFlintlockClient
		cfg        *config.Config
	)

	BeforeEach(func() {
		mockClient = new(fakeclient.FakeFlintlockClient)
		builderFunc := func(string, string) (client.FlintlockClient, error) {
			return mockClient, nil
		}
		cfg = &config.Config{
			ClientConfig: config.ClientConfig{
				ClientBuilderFunc: builderFunc,
			},
			MvmName:      name,
			MvmNamespace: namespace,
		}
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

		mockClient.CreateReturns(microVm, nil)
		err := command.CreateFn(cfg)
		Expect(err).ToNot(HaveOccurred())

		input := mockClient.CreateArgsForCall(0)
		Expect(input.Id).To(Equal(name))
		Expect(input.Namespace).To(Equal(namespace))

		var userData userdata.UserData
		data, err := base64.StdEncoding.DecodeString(input.Metadata["user-data"])
		Expect(err).ToNot(HaveOccurred())
		Expect(yaml.Unmarshal(data, &userData)).To(Succeed())
		Expect(userData.HostName).To(Equal(name))
		Expect(userData.Users[0].Name).To(Equal("root"))

		var metaData instance.Metadata
		data, err = base64.StdEncoding.DecodeString(input.Metadata["meta-data"])
		Expect(err).ToNot(HaveOccurred())
		Expect(yaml.Unmarshal(data, &metaData)).To(Succeed())
		Expect(metaData["instance_id"]).To(Equal(fmt.Sprintf("%s/%s", namespace, name)))
		Expect(metaData["local_hostname"]).To(Equal(name))
		Expect(metaData["platform"]).To(Equal("liquid_metal"))
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

			JustBeforeEach(func() {
				cfg.SSHKeyPath = keyFile.Name()
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

				mockClient.CreateReturns(microVm, nil)
				err := command.CreateFn(cfg)
				Expect(err).ToNot(HaveOccurred())

				input := mockClient.CreateArgsForCall(0)
				Expect(input.Id).To(Equal(name))
				Expect(input.Namespace).To(Equal(namespace))
				var user userdata.UserData
				userData, err := base64.StdEncoding.DecodeString(input.Metadata["user-data"])
				Expect(err).ToNot(HaveOccurred())
				Expect(yaml.Unmarshal(userData, &user)).To(Succeed())
				Expect(user.Users[0].Name).To(Equal("root"))
				Expect(user.Users[0].SSHAuthorizedKeys[0]).To(Equal("this is a test key woohoo"))

				Expect(mockClient.CreateCallCount()).To(Equal(1))
			})
		})

		Context("when file does not exist", func() {
			JustBeforeEach(func() {
				cfg.SSHKeyPath = "foo.txt"
			})

			It("returns an error", func() {
				err := command.CreateFn(cfg)
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

		BeforeEach(func() {
			cfg.JSONFile = jsonSpec
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

			mockClient.CreateReturns(microVm, nil)
			err := command.CreateFn(cfg)
			Expect(err).ToNot(HaveOccurred())

			input := mockClient.CreateArgsForCall(0)
			Expect(input.Id).To(Equal(name))
			Expect(input.Namespace).To(Equal(namespace))

			Expect(mockClient.CreateCallCount()).To(Equal(1))
		})

		Context("when file does not exist", func() {
			BeforeEach(func() {
				cfg.JSONFile = "./../../example1.json"
			})

			It("returns an error", func() {
				err := command.CreateFn(cfg)
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
			})
		})
	})
})
