package command_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"

	"github.com/warehouse-13/hammertime/pkg/client/fakeclient"
	"github.com/warehouse-13/hammertime/pkg/command"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/utils"
)

func Test_ListFn(t *testing.T) {
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

	resp := listResponse(2, testName, testNamespace)
	mockClient.ListReturns(resp, nil)
	g.Expect(command.ListFn(w, cfg)).To(Succeed())

	inName, inNamespace := mockClient.ListArgsForCall(0)
	g.Expect(inName).To(Equal(testName))
	g.Expect(inNamespace).To(Equal(testNamespace))

	out := &v1alpha1.ListMicroVMsResponse{}
	g.Expect(json.Unmarshal(buf.Bytes(), out)).To(Succeed())

	g.Expect(out.Microvm).To(Equal(resp.Microvm))
	g.Expect(out.Microvm).To(HaveLen(2))
}

func Test_ListFn_clientFails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, nil),
		},
	}

	mockClient.ListReturns(nil, errors.New("error"))
	g.Expect(command.ListFn(utils.NewWriter(nil), cfg)).NotTo(Succeed())
}

func Test_ListFn_clientBuilderFails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, errors.New("unusable")),
		},
	}

	g.Expect(command.ListFn(utils.NewWriter(nil), cfg)).NotTo(Succeed())
}
