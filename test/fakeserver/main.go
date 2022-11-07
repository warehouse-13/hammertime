package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
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

func main() {
	l, err := net.Listen("tcp", "localhost:9090")
	if err != nil {
		fmt.Println("Failed to listen on localhost:9090", err)
		os.Exit(1)
	}

	s := &fakeServer{}
	token := os.Getenv("AUTH_TOKEN")
	grpcServer := grpc.NewServer(withOpts(token)...)
	mvmv1.RegisterMicroVMServer(grpcServer, s)

	if err := grpcServer.Serve(l); err != nil {
		fmt.Println("Failed to start gRPC server", err)
		os.Exit(1)
	}
}

type fakeServer struct {
	savedSpecs []*types.MicroVMSpec
}

func (s *fakeServer) CreateMicroVM(
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

func (s *fakeServer) DeleteMicroVM(ctx context.Context, req *mvmv1.DeleteMicroVMRequest) (*emptypb.Empty, error) {
	for i, spec := range s.savedSpecs {
		if *spec.Uid == req.Uid {
			s.savedSpecs[i] = s.savedSpecs[len(s.savedSpecs)-1]
		}
	}

	s.savedSpecs = s.savedSpecs[:len(s.savedSpecs)-1]

	return &emptypb.Empty{}, nil
}

func (s *fakeServer) GetMicroVM(ctx context.Context, req *mvmv1.GetMicroVMRequest) (*mvmv1.GetMicroVMResponse, error) {
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

func (s *fakeServer) ListMicroVMs(
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

func (s *fakeServer) ListMicroVMsStream(
	req *mvmv1.ListMicroVMsRequest,
	streamServer mvmv1.MicroVM_ListMicroVMsStreamServer,
) error {
	return nil
}

func withOpts(authToken string) []grpc.ServerOption {
	if authToken != "" {
		fmt.Println("basic authentication is enabled")

		return []grpc.ServerOption{
			grpc.StreamInterceptor(grpc_mw.ChainStreamServer(
				grpc_auth.StreamServerInterceptor(basicAuthFunc(authToken)),
			)),
			grpc.UnaryInterceptor(grpc_mw.ChainUnaryServer(
				grpc_auth.UnaryServerInterceptor(basicAuthFunc(authToken)),
			)),
		}
	}

	fmt.Println("authentication is DISABLED")

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
