package command_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/warehouse-13/hammertime/pkg/client/fakeclient"
	"github.com/warehouse-13/hammertime/pkg/command"
	"github.com/warehouse-13/hammertime/pkg/config"
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

	mockClient.ListReturns(listResponse(1, testName, testNamespace), nil)
	g.Expect(command.ListFn(cfg)).To(Succeed())

	inName, inNamespace := mockClient.ListArgsForCall(0)
	g.Expect(inName).To(Equal(testName))
	g.Expect(inNamespace).To(Equal(testNamespace))
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
	g.Expect(command.ListFn(cfg)).NotTo(Succeed())
}

func Test_ListFn_clientBuilderFails(t *testing.T) {
	g := NewWithT(t)

	mockClient := new(fakeclient.FakeFlintlockClient)
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: testClient(mockClient, errors.New("unusable")),
		},
	}

	g.Expect(command.ListFn(cfg)).NotTo(Succeed())
}
