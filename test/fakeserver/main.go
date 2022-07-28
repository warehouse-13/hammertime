package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	uuid "github.com/nu7hatch/gouuid"
	mvmv1 "github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"
	"google.golang.org/grpc"
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

	grpcServer := grpc.NewServer()
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
