package command_test

import (
	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"
)

func testClient(c client.FlintlockClient, err error) func(string, string) (client.FlintlockClient, error) {
	return func(string, string) (client.FlintlockClient, error) {
		return c, err
	}
}

func createResponse(name, namespace string) *v1alpha1.CreateMicroVMResponse {
	return &v1alpha1.CreateMicroVMResponse{
		Microvm: &types.MicroVM{
			Spec: &types.MicroVMSpec{
				Id:        name,
				Namespace: namespace,
			},
		},
	}
}

func listResponse(name, namespace string) *v1alpha1.ListMicroVMsResponse {
	return &v1alpha1.ListMicroVMsResponse{
		Microvm: []*types.MicroVM{{
			Spec: &types.MicroVMSpec{
				Id:        name,
				Namespace: namespace,
			},
		}},
	}
}
