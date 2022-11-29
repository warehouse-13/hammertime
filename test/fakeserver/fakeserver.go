package fakeserver

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"

	grpc_mw "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	uuid "github.com/nu7hatch/gouuid"
	mvmv1 "github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/warehouse-13/hammertime/pkg/utils"
)

// Cleanup implements a function to tear down any fake server resources.
type Cleanup func() error

func New() *FakeServer {
	return &FakeServer{}
}

type FakeServer struct {
	savedSpecs []*types.MicroVMSpec
	cleanup    Cleanup
}

// Start creates a new real server to respond to gRPC requests from the client.
// Using a buffcon would be better, but this only works if you are creating
// the client programatically in the tests. In our case we are passing the server
// address as a flag to the program. The client therefore errors because it
// does not recognise the buffer address as a valid target.
// The fake server has additional methods which allow for manipulation of data.
func (s *FakeServer) Start(token string) string {
	l, err := net.Listen("tcp", ":")
	if err != nil {
		fmt.Println("Failed to start fake listener", err)
	}

	grpcServer := grpc.NewServer(withOpts(token)...)
	mvmv1.RegisterMicroVMServer(grpcServer, s)

	go func() {
		if err := grpcServer.Serve(l); err != nil {
			fmt.Println("Failed to start fake gRPC server", err)
		}
	}()

	s.cleanup = func() error {
		if err := l.Close(); err != nil {
			return err
		}

		grpcServer.Stop()

		return nil
	}

	return l.Addr().String()
}

func (s *FakeServer) Stop() error {
	if s.cleanup != nil {
		return s.cleanup()
	}

	return nil
}

func (s *FakeServer) CreateMicroVM(
	ctx context.Context,
	req *mvmv1.CreateMicroVMRequest,
) (*mvmv1.CreateMicroVMResponse, error) {
	spec := req.Microvm

	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	spec.Uid = utils.PointyString(uid.String())

	s.savedSpecs = append(s.savedSpecs, spec)

	return &mvmv1.CreateMicroVMResponse{
		Microvm: &types.MicroVM{
			Version: 0,
			Spec:    spec,
			Status:  &types.MicroVMStatus{},
		},
	}, nil
}

func (s *FakeServer) DeleteMicroVM(ctx context.Context, req *mvmv1.DeleteMicroVMRequest) (*emptypb.Empty, error) {
	for i, spec := range s.savedSpecs {
		if *spec.Uid == req.Uid {
			s.savedSpecs[i] = s.savedSpecs[len(s.savedSpecs)-1]
		}
	}

	s.savedSpecs = s.savedSpecs[:len(s.savedSpecs)-1]

	return &emptypb.Empty{}, nil
}

func (s *FakeServer) GetMicroVM(ctx context.Context, req *mvmv1.GetMicroVMRequest) (*mvmv1.GetMicroVMResponse, error) {
	var requestSpec *types.MicroVMSpec

	for _, spec := range s.savedSpecs {
		if *spec.Uid == req.Uid {
			requestSpec = spec
		}
	}

	if requestSpec == nil {
		return nil, errors.New("rpc error: OHH WHAT A DISASTER")
	}

	return &mvmv1.GetMicroVMResponse{
		Microvm: &types.MicroVM{
			Version: 0,
			Spec:    requestSpec,
			Status: &types.MicroVMStatus{
				State: types.MicroVMStatus_CREATED,
			},
		},
	}, nil
}

func (s *FakeServer) ListMicroVMs(
	ctx context.Context,
	req *mvmv1.ListMicroVMsRequest,
) (*mvmv1.ListMicroVMsResponse, error) {
	microvms := []*types.MicroVM{}

	for _, spec := range s.savedSpecs {
		if shouldReturn(spec, req.Name, req.Namespace) {
			m := &types.MicroVM{
				Version: 0,
				Spec:    spec,
				Status: &types.MicroVMStatus{
					State: types.MicroVMStatus_CREATED,
				},
			}
			microvms = append(microvms, m)
		}
	}

	return &mvmv1.ListMicroVMsResponse{
		Microvm: microvms,
	}, nil
}

func shouldReturn(spec *types.MicroVMSpec, name *string, namespace string) bool {
	if spec.Namespace == namespace && spec.Id == *name {
		return true
	}

	if spec.Namespace == namespace && *name == "" {
		return true
	}

	return namespace == ""
}

func (s *FakeServer) ListMicroVMsStream(
	req *mvmv1.ListMicroVMsRequest,
	streamServer mvmv1.MicroVM_ListMicroVMsStreamServer,
) error {
	return nil
}

func withOpts(authToken string) []grpc.ServerOption {
	if authToken != "" {
		return []grpc.ServerOption{
			grpc.StreamInterceptor(grpc_mw.ChainStreamServer(
				grpc_auth.StreamServerInterceptor(basicAuthFunc(authToken)),
			)),
			grpc.UnaryInterceptor(grpc_mw.ChainUnaryServer(
				grpc_auth.UnaryServerInterceptor(basicAuthFunc(authToken)),
			)),
		}
	}

	return []grpc.ServerOption{
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	}
}

func basicAuthFunc(setServerToken string) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := grpc_auth.AuthFromMD(ctx, "basic")
		if err != nil {
			return nil, fmt.Errorf("could not extract token from request header: %w", err)
		}

		if err := validateBasicAuthToken(token, setServerToken); err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
		}

		return ctx, nil
	}
}

func validateBasicAuthToken(clientToken string, serverToken string) error {
	data := base64.StdEncoding.EncodeToString([]byte(serverToken))

	if strings.Compare(clientToken, string(data)) != 0 {
		return errors.New("tokens do not match")
	}

	return nil
}
