package command_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"
	"k8s.io/utils/pointer"

	"github.com/warehouse-13/hammertime/pkg/client/fakeclient"
	"github.com/warehouse-13/hammertime/pkg/command"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/utils"
)

func Test_GetFn(t *testing.T) {
	g := NewWithT(t)

	var (
		testName      = "foo"
		testNamespace = "bar"
		testUid       = "abc123"
	)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		UUID: testUid,
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	resp := getResponse(testName, testNamespace, testUid)
	mockClient.GetReturns(resp, nil)
	g.Expect(command.GetFn(w, cfg)).To(Succeed())

	inUid := mockClient.GetArgsForCall(0)
	g.Expect(inUid).To(Equal(testUid))

	out := &v1alpha1.GetMicroVMResponse{}
	g.Expect(json.Unmarshal(buf.Bytes(), out)).To(Succeed())

	g.Expect(out.Microvm).To(Equal(resp.Microvm))
}

func Test_GetFn_state(t *testing.T) {
	g := NewWithT(t)

	var (
		testName      = "foo"
		testNamespace = "bar"
		testUid       = "abc123"
	)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		UUID:  testUid,
		State: true,
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	mockClient.GetReturns(getResponse(testName, testNamespace, testUid), nil)
	g.Expect(command.GetFn(w, cfg)).To(Succeed())

	inUid := mockClient.GetArgsForCall(0)
	g.Expect(inUid).To(Equal(testUid))

	g.Expect(buf.String()).To(Equal("CREATED\n"))
}

func Test_GetFn_uidNotSet(t *testing.T) {
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
	g.Expect(command.GetFn(w, cfg)).To(Succeed())

	g.Expect(mockClient.GetCallCount()).To(BeZero())
	inName, inNamespace := mockClient.ListArgsForCall(0)
	g.Expect(inName).To(Equal(testName))
	g.Expect(inNamespace).To(Equal(testNamespace))

	out := &types.MicroVM{}
	g.Expect(json.Unmarshal(buf.Bytes(), out)).To(Succeed())

	g.Expect(out).To(Equal(resp.Microvm[0]))
}

func Test_GetFn_uidNotSet_state(t *testing.T) {
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
		State:        true,
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	mockClient.ListReturns(listResponse(1, testName, testName), nil)
	g.Expect(command.GetFn(w, cfg)).To(Succeed())

	g.Expect(mockClient.GetCallCount()).To(BeZero())
	inName, inNamespace := mockClient.ListArgsForCall(0)
	g.Expect(inName).To(Equal(testName))
	g.Expect(inNamespace).To(Equal(testNamespace))

	g.Expect(buf.String()).To(Equal("CREATED\n"))
}

func Test_GetFn_uidNotSet_multipleMatches(t *testing.T) {
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

	mockClient.ListReturns(listResponse(2, testName, testNamespace), nil)
	g.Expect(command.GetFn(w, cfg)).To(Succeed())

	g.Expect(mockClient.GetCallCount()).To(BeZero())
	inName, inNamespace := mockClient.ListArgsForCall(0)
	g.Expect(inName).To(Equal(testName))
	g.Expect(inNamespace).To(Equal(testNamespace))

	g.Expect(buf.String()).To(ContainSubstring(fmt.Sprintf("2 MicroVMs found under %s/%s", testNamespace, testName)))
}

func Test_GetFn_withFile(t *testing.T) {
	g := NewWithT(t)

	var (
		testName      = "fname"
		testNamespace = "fns"
		testUid       = "abc123"
	)

	spec := &types.MicroVMSpec{
		Id:        testName,
		Namespace: testNamespace,
		Uid:       pointer.String(testUid),
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

	resp := getResponse(testName, testNamespace, testUid)
	mockClient.GetReturns(resp, nil)
	g.Expect(command.GetFn(w, cfg)).To(Succeed())

	inUid := mockClient.GetArgsForCall(0)
	g.Expect(inUid).To(Equal(testUid))

	out := &v1alpha1.GetMicroVMResponse{}
	g.Expect(json.Unmarshal(buf.Bytes(), out)).To(Succeed())

	g.Expect(out.Microvm).To(Equal(resp.Microvm))
}

func Test_GetFn_withFile_fails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		JSONFile: "noexist",
	}

	g.Expect(command.GetFn(utils.NewWriter(nil), cfg)).NotTo(Succeed())
}

func Test_GetFn_nothingFound(t *testing.T) {
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

	mockClient.ListReturns(listResponse(0, "", ""), nil)
	err := command.GetFn(utils.NewWriter(nil), cfg)
	g.Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("MicroVM %s/%s not found", testNamespace, testName))))
}

func Test_GetFn_clientFails(t *testing.T) {
	g := NewWithT(t)

	var testUid = "abc123"

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
		UUID: testUid,
	}

	mockClient.GetReturns(nil, errors.New("error"))
	g.Expect(command.GetFn(utils.NewWriter(nil), cfg)).NotTo(Succeed())
}

func Test_GetFn_clientBuilderFails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, errors.New("unusable")),
		},
	}

	g.Expect(command.GetFn(utils.NewWriter(nil), cfg)).NotTo(Succeed())
}
