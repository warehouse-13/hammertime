package command_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"
	"k8s.io/utils/pointer"

	"github.com/warehouse-13/hammertime/pkg/client/fakeclient"
	"github.com/warehouse-13/hammertime/pkg/command"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/utils"
)

func Test_DeleteFn(t *testing.T) {
	g := NewWithT(t)

	var testUid = "123abc"

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		UUID: testUid,
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	mockClient.DeleteReturns(deleteResponse(), nil)
	g.Expect(command.DeleteFn(w, cfg)).To(Succeed())

	input := mockClient.DeleteArgsForCall(0)
	g.Expect(input).To(Equal(testUid))

	g.Expect(buf.String()).To(Equal("{}\n"))
}

func Test_DeleteFn_silent(t *testing.T) {
	g := NewWithT(t)

	var testUid = "123abc"

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		UUID:   testUid,
		Silent: true,
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	mockClient.DeleteReturns(deleteResponse(), nil)
	g.Expect(command.DeleteFn(w, cfg)).To(Succeed())

	input := mockClient.DeleteArgsForCall(0)
	g.Expect(input).To(Equal(testUid))

	g.Expect(buf.String()).To(BeEmpty())
}

func Test_DeleteFn_noUid_noDeleteAll_nameNotSet(t *testing.T) {
	g := NewWithT(t)

	var testNamespace = "bar"

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		MvmNamespace: testNamespace,
	}

	err := command.DeleteFn(utils.NewWriter(nil), cfg)
	g.Expect(err).To(MatchError("required: --namespace, --name"))
}

func Test_DeleteFn_noUid_noDeleteAll_namespaceNotSet(t *testing.T) {
	g := NewWithT(t)

	var testName = "foo"

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		MvmName: testName,
	}

	err := command.DeleteFn(utils.NewWriter(nil), cfg)
	g.Expect(err).To(MatchError("required: --namespace, --name"))
}

func Test_DeleteFn_withFile(t *testing.T) {
	g := NewWithT(t)

	var testUid = "123abc"

	spec := &types.MicroVMSpec{
		Uid: pointer.String(testUid),
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

	mockClient.DeleteReturns(deleteResponse(), nil)
	g.Expect(command.DeleteFn(w, cfg)).To(Succeed())

	input := mockClient.DeleteArgsForCall(0)
	g.Expect(input).To(Equal(testUid))

	g.Expect(buf.String()).To(Equal("{}\n"))
}

func Test_DeleteFn_withFile_fails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		JSONFile: "noexist",
	}

	g.Expect(command.DeleteFn(utils.NewWriter(nil), cfg)).NotTo(Succeed())
}

func Test_DeleteFn_noUid_noDeleteAll_oneMatch(t *testing.T) {
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

	resp := listResponse(1, testName, testName)
	mockClient.ListReturns(resp, nil)
	mockClient.DeleteReturns(deleteResponse(), nil)

	g.Expect(command.DeleteFn(w, cfg)).To(Succeed())

	input := mockClient.DeleteArgsForCall(0)
	g.Expect(input).To(Equal(*resp.Microvm[0].Spec.Uid))

	g.Expect(buf.String()).To(Equal("{}\n"))
}

func Test_DeleteFn_noUid_noDeleteAll_multipleMatches(t *testing.T) {
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

	resp := listResponse(2, testName, testName)
	mockClient.ListReturns(resp, nil)

	g.Expect(command.DeleteFn(w, cfg)).To(Succeed())

	g.Expect(mockClient.DeleteCallCount()).To(BeZero())

	g.Expect(buf.String()).To(ContainSubstring(fmt.Sprintf("2 MicroVMs found under %s/%s:", testNamespace, testName)))
}

func Test_DeleteFn_noUid_deleteAll(t *testing.T) {
	g := NewWithT(t)

	var (
		testName      = "foo"
		testNamespace = "bar"
		mvmCount      = 2
	)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		MvmName:      testName,
		MvmNamespace: testNamespace,
		DeleteAll:    true,
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	resp := listResponse(mvmCount, testName, testName)
	mockClient.ListReturns(resp, nil)
	mockClient.DeleteReturns(deleteResponse(), nil)

	g.Expect(command.DeleteFn(w, cfg)).To(Succeed())

	g.Expect(mockClient.DeleteCallCount()).To(Equal(mvmCount))

	input := mockClient.DeleteArgsForCall(0)
	g.Expect(input).To(Equal(*resp.Microvm[0].Spec.Uid))
	input = mockClient.DeleteArgsForCall(1)
	g.Expect(input).To(Equal(*resp.Microvm[1].Spec.Uid))

	g.Expect(buf.String()).To(Equal("{}\n{}\n"))
}

func Test_DeleteFn_noUid_deleteAll_silent(t *testing.T) {
	g := NewWithT(t)

	var (
		testName      = "foo"
		testNamespace = "bar"
		mvmCount      = 2
	)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		MvmName:      testName,
		MvmNamespace: testNamespace,
		DeleteAll:    true,
		Silent:       true,
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	resp := listResponse(mvmCount, testName, testName)
	mockClient.ListReturns(resp, nil)
	mockClient.DeleteReturns(deleteResponse(), nil)

	g.Expect(command.DeleteFn(w, cfg)).To(Succeed())

	g.Expect(mockClient.DeleteCallCount()).To(Equal(mvmCount))

	input := mockClient.DeleteArgsForCall(0)
	g.Expect(input).To(Equal(*resp.Microvm[0].Spec.Uid))
	input = mockClient.DeleteArgsForCall(1)
	g.Expect(input).To(Equal(*resp.Microvm[1].Spec.Uid))

	g.Expect(buf.String()).To(BeEmpty())
}

func Test_DeleteFn_clientFails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
	}

	mockClient.DeleteReturns(nil, errors.New("error"))
	g.Expect(command.DeleteFn(utils.NewWriter(nil), cfg)).NotTo(Succeed())
}

func Test_DeleteFn_clientBuilderFails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, errors.New("unusable")),
		},
	}

	g.Expect(command.DeleteFn(utils.NewWriter(nil), cfg)).NotTo(Succeed())
}
