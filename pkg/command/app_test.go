package command_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/warehouse-13/safety"
	"github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"
	"google.golang.org/grpc"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/command"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/dialler"
	"github.com/warehouse-13/hammertime/pkg/utils"
)

func Test_CRUD_noBasicAuth_noTLS(t *testing.T) {
	g := NewWithT(t)

	fakeserver := safety.New()
	dialer := fakeserver.StartBuf("")

	t.Cleanup(func() {
		fakeserver.Stop()
	})

	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: cl(dialer),
		},
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	g.Expect(command.CreateFn(w, cfg)).To(Succeed())

	out := &v1alpha1.CreateMicroVMResponse{}
	g.Expect(json.Unmarshal(buf.Bytes(), out)).To(Succeed())

	cfg.UUID = *out.Microvm.Spec.Uid

	g.Expect(command.GetFn(w, cfg)).To(Succeed())
	g.Expect(command.ListFn(w, cfg)).To(Succeed())
	g.Expect(command.DeleteFn(w, cfg)).To(Succeed())
}

func Test_CRUD_basicAuth_noTLS(t *testing.T) {
	g := NewWithT(t)

	basicAuthToken := "secret"

	fakeserver := safety.New()
	dialer := fakeserver.StartBuf(basicAuthToken)

	t.Cleanup(func() {
		fakeserver.Stop()
	})

	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: cl(dialer),
		},
		Token: basicAuthToken,
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	g.Expect(command.CreateFn(w, cfg)).To(Succeed())

	out := &v1alpha1.CreateMicroVMResponse{}
	g.Expect(json.Unmarshal(buf.Bytes(), out)).To(Succeed())

	cfg.UUID = *out.Microvm.Spec.Uid

	g.Expect(command.GetFn(w, cfg)).To(Succeed())
	g.Expect(command.ListFn(w, cfg)).To(Succeed())
	g.Expect(command.DeleteFn(w, cfg)).To(Succeed())
}

func Test_basicAuth_failsWithNoClientToken(t *testing.T) {
	g := NewWithT(t)

	basicAuthToken := "secret"

	fakeserver := safety.New()
	dialer := fakeserver.StartBuf(basicAuthToken)

	t.Cleanup(func() {
		fakeserver.Stop()
	})

	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: cl(dialer),
		},
		Token: "",
	}

	buf := &bytes.Buffer{}
	w := utils.NewWriter(buf)

	g.Expect(command.CreateFn(w, cfg)).To(MatchError(ContainSubstring("unauthenticated")))
}

func cl(dialer func(context.Context, string) (net.Conn, error)) func(string, string) (client.FlintlockClient, error) {
	return func(string, token string) (client.FlintlockClient, error) {
		opt := []grpc.DialOption{grpc.WithContextDialer(dialer)}
		conn, err := dialler.New("bufnet", token, opt)
		if err != nil {
			return nil, err
		}

		return &client.Client{
			MicroVMClient: v1alpha1.NewMicroVMClient(conn),
			Conn:          conn,
		}, nil
	}
}
