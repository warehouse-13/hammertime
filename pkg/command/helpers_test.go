package command_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"
	"k8s.io/utils/pointer"
)

func testClient(c client.FlintlockClient, err error) func(string, string) (client.FlintlockClient, error) {
	return func(string, string) (client.FlintlockClient, error) {
		return c, err
	}
}

func writeFile(spec *types.MicroVMSpec) (*os.File, error) {
	tempFile, err := ioutil.TempFile("", "getfn_test")
	if err != nil {
		return nil, err
	}

	dat, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(tempFile.Name(), dat, 0755); err != nil {
		return nil, err
	}

	return tempFile, nil
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

func getResponse(name, namespace, uid string) *v1alpha1.GetMicroVMResponse {
	return &v1alpha1.GetMicroVMResponse{
		Microvm: &types.MicroVM{
			Spec: &types.MicroVMSpec{
				Id:        name,
				Namespace: namespace,
				Uid:       pointer.String(uid),
			},
			Status: &types.MicroVMStatus{
				State: types.MicroVMStatus_CREATED,
			},
		},
	}
}

func listResponse(count int, name, namespace string) *v1alpha1.ListMicroVMsResponse {
	mvms := []*types.MicroVM{}

	for i := 0; i < count; i++ {
		mvm := &types.MicroVM{
			Spec: &types.MicroVMSpec{
				Id:        name,
				Namespace: namespace,
				Uid:       pointer.String(randomString(10)),
			},
			Status: &types.MicroVMStatus{
				State: types.MicroVMStatus_CREATED,
			},
		}

		mvms = append(mvms, mvm)
	}

	return &v1alpha1.ListMicroVMsResponse{
		Microvm: mvms,
	}
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}
