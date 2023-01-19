package command_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"

	"github.com/warehouse-13/hammertime/pkg/client/fakeclient"
	"github.com/warehouse-13/hammertime/pkg/command"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/utils"
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

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	resp := createResponse(testName, testNamespace)
	mockClient.CreateReturns(resp, nil)
	g.Expect(command.CreateFn(w, cfg)).To(Succeed())

	input := mockClient.CreateArgsForCall(0)
	g.Expect(input.Id).To(Equal(testName))
	g.Expect(input.Namespace).To(Equal(testNamespace))

	out := &v1alpha1.CreateMicroVMResponse{}
	g.Expect(json.Unmarshal(buf.Bytes(), out)).To(Succeed())

	g.Expect(out.Microvm).To(Equal(resp.Microvm))
}

func Test_CreateFn_silent(t *testing.T) {
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
		Silent:       true,
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	mockClient.CreateReturns(createResponse(testName, testNamespace), nil)
	g.Expect(command.CreateFn(w, cfg)).To(Succeed())

	input := mockClient.CreateArgsForCall(0)
	g.Expect(input.Id).To(Equal(testName))
	g.Expect(input.Namespace).To(Equal(testNamespace))

	g.Expect(buf.String()).To(BeEmpty())
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
	g.Expect(command.CreateFn(utils.NewWriter(nil), cfg)).NotTo(Succeed())
}

func Test_CreateFn_clientBuilderFails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, errors.New("unusable")),
		},
	}

	g.Expect(command.CreateFn(utils.NewWriter(nil), cfg)).NotTo(Succeed())
}

func Test_CreateFn_withFile(t *testing.T) {
	g := NewWithT(t)

	var (
		testName      = "fname"
		testNamespace = "fns"
	)

	spec := &types.MicroVMSpec{
		Id:        testName,
		Namespace: testNamespace,
	}

	tempFile, err := writeFile(spec)
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		g.Expect(os.RemoveAll(tempFile.Name())).To(Succeed())
	})

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		JSONFile: tempFile.Name(),
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	resp := createResponse(testName, testNamespace)
	mockClient.CreateReturns(resp, nil)
	g.Expect(command.CreateFn(w, cfg)).To(Succeed())

	input := mockClient.CreateArgsForCall(0)
	g.Expect(input.Id).To(Equal(testName))
	g.Expect(input.Namespace).To(Equal(testNamespace))

	out := &v1alpha1.CreateMicroVMResponse{}
	g.Expect(json.Unmarshal(buf.Bytes(), out)).To(Succeed())

	g.Expect(out.Microvm).To(Equal(resp.Microvm))
}

func Test_CreateFn_withFile_fails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		JSONFile: "noexist",
	}

	g.Expect(command.CreateFn(utils.NewWriter(nil), cfg)).NotTo(Succeed())
}
