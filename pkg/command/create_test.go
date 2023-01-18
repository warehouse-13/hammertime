package command_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/client/fakeclient"
	"github.com/warehouse-13/hammertime/pkg/command"
	"github.com/warehouse-13/hammertime/pkg/config"
)

func Test_CreateFn(t *testing.T) {
	g := NewWithT(t)

	var (
		testName      = "foo"
		testNamespace = "bar"
	)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		MvmName:      testName,
		MvmNamespace: testNamespace,
	}

	mockClient.CreateReturns(response(testName, testNamespace), nil)
	g.Expect(command.CreateFn(cfg)).To(Succeed())

	input := mockClient.CreateArgsForCall(0)
	g.Expect(input.Id).To(Equal(testName))
	g.Expect(input.Namespace).To(Equal(testNamespace))
}

func Test_CreateFn_clientFails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
	}

	mockClient.CreateReturns(nil, errors.New("error"))
	g.Expect(command.CreateFn(cfg)).NotTo(Succeed())
}

func Test_CreateFn_clientBuilderFails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, errors.New("unusable")),
		},
	}

	g.Expect(command.CreateFn(cfg)).NotTo(Succeed())
}

func Test_CreateFn_withFile(t *testing.T) {
	g := NewWithT(t)

	var (
		testName      = "fname"
		testNamespace = "fns"
	)

	tempFile, err := ioutil.TempFile("", "createfn_test")
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		g.Expect(os.RemoveAll(tempFile.Name())).To(Succeed())
	})

	spec := &types.MicroVMSpec{
		Id:        testName,
		Namespace: testNamespace,
	}

	dat, err := json.Marshal(spec)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(ioutil.WriteFile(tempFile.Name(), dat, 0755)).To(Succeed())

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		MvmName:      testName,
		MvmNamespace: testNamespace,
		JSONFile:     tempFile.Name(),
	}

	mockClient.CreateReturns(response(testName, testNamespace), nil)
	g.Expect(command.CreateFn(cfg)).To(Succeed())

	input := mockClient.CreateArgsForCall(0)
	g.Expect(input.Id).To(Equal(testName))
	g.Expect(input.Namespace).To(Equal(testNamespace))
}

func Test_CreateFn_withFile_fails(t *testing.T) {
	g := NewWithT(t)

	var (
		testName      = "fname"
		testNamespace = "fns"
	)

	tempFile, err := ioutil.TempFile("", "createfn_test")
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		g.Expect(os.RemoveAll(tempFile.Name())).To(Succeed())
	})

	g.Expect(ioutil.WriteFile(tempFile.Name(), []byte("foo"), 0755)).To(Succeed())

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		MvmName:      testName,
		MvmNamespace: testNamespace,
		JSONFile:     tempFile.Name(),
	}

	g.Expect(command.CreateFn(cfg)).NotTo(Succeed())
}

func testClient(c client.FlintlockClient, err error) func(string, string) (client.FlintlockClient, error) {
	return func(string, string) (client.FlintlockClient, error) {
		return c, err
	}
}

func response(name, namespace string) *v1alpha1.CreateMicroVMResponse {
	return &v1alpha1.CreateMicroVMResponse{
		Microvm: &types.MicroVM{
			Spec: &types.MicroVMSpec{
				Id:        name,
				Namespace: namespace,
			},
		},
	}
}
