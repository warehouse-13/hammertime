package microvm

import (
	"fmt"
	"log"

	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks/flintlock/api/types"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/defaults"
	"github.com/warehouse-13/hammertime/pkg/microvm/data"
	"github.com/warehouse-13/hammertime/pkg/utils"
)

// MicroVMManager .
type MicroVMManager struct { //nolint: revive // no worries
	c client.Client
}

// NewManager .
// have this return an interface so I can test the action funcs?
func NewManager(c client.Client) MicroVMManager {
	return MicroVMManager{c}
}

// Create creates a new Microvm.
// If a json file is set, creates one based on the given spec, otherwise creates
// a microvm with stanard config.
func (m MicroVMManager) Create(cfg *config.Config) (*v1alpha1.CreateMicroVMResponse, error) {
	var (
		mvm *types.MicroVMSpec
		err error
	)

	if utils.IsSet(cfg.JSONFile) {
		mvm, err = utils.LoadSpecFromFile(cfg.JSONFile)
		if err != nil {
			return nil, err
		}
	} else {
		mvm, err = newMicroVM(cfg.MvmName, cfg.MvmNamespace, cfg.SSHKeyPath)
		if err != nil {
			return nil, err
		}
	}

	return m.c.Create(mvm)
}

// TODO: add tests as part of #23.
// Get gets a Microvm.
// If a Json file is set on the config, the values in that file will be used.
// If a UUID is set or found, that is the mvm which will be fetched.
// If State is set, then just the state of the found mvm will be returned.
// If a namespace/name are set, then a List will be called. If there is only one mvm
// in that namespace/name, then it will be fetched. If more than one is found, the
// found ones will be printed out.
func (m MicroVMManager) Get(cfg *config.Config) (interface{}, error) {
	if utils.IsSet(cfg.JSONFile) {
		var err error

		cfg.UUID, cfg.MvmName, cfg.MvmNamespace, err = utils.ProcessFile(cfg.JSONFile)
		if err != nil {
			return nil, err
		}
	}

	if utils.IsSet(cfg.UUID) {
		res, err := m.c.Get(cfg.UUID)
		if err != nil {
			return nil, err
		}

		if cfg.State {
			return fmt.Sprint(res.Microvm.Status.State), nil
		}

		return res, nil
	}

	res, err := m.c.List(cfg.MvmName, cfg.MvmNamespace)
	if err != nil {
		return nil, err
	}

	if len(res.Microvm) > 1 {
		log.Printf("%d MicroVMs found under %s/%s:\n", len(res.Microvm), cfg.MvmNamespace, cfg.MvmName)

		var found []string

		for _, mvm := range res.Microvm {
			found = append(found, *mvm.Spec.Uid)
		}

		return found, nil
	}

	if len(res.Microvm) == 1 {
		if cfg.State {
			return fmt.Sprint(res.Microvm[0].Status.State), nil
		}

		return res.Microvm[0], nil
	}

	return nil, fmt.Errorf("MicroVM %s/%s not found", cfg.MvmName, cfg.MvmNamespace)
}

// List fetches all Microvms.
// TODO: add tests as part of #23.
func (m MicroVMManager) List(cfg *config.Config) (*v1alpha1.ListMicroVMsResponse, error) {
	return m.c.List(cfg.MvmName, cfg.MvmNamespace)
}

// TODO: add tests as part of #23.
// Delete deletes an existing Microvm.
// If a Json file is set on the config, the values in that file will be used.
// If a UUID is set or found, that is the mvm which will be deleted.
// If DeleteAll is set, then all found microvms will be deleted.
// If a namespace/name are set, then all mvms in that namespace/name will be deleted.
//nolint: cyclop // we are refactoring this func
func (m MicroVMManager) Delete(cfg *config.Config) (interface{}, error) {
	if utils.IsSet(cfg.JSONFile) {
		var err error

		cfg.UUID, cfg.MvmName, cfg.MvmNamespace, err = utils.ProcessFile(cfg.JSONFile)
		if err != nil {
			return nil, err
		}
	}

	if utils.IsSet(cfg.UUID) {
		return m.c.Delete(cfg.UUID)
	}

	if cfg.DeleteAll {
		if utils.IsSet(cfg.MvmName) && !utils.IsSet(cfg.MvmNamespace) {
			return nil, fmt.Errorf("required: --namespace")
		}
	} else {
		if utils.IsSet(cfg.MvmName) && !utils.IsSet(cfg.MvmNamespace) {
			return nil, fmt.Errorf("required: --namespace")
		}
		if !utils.IsSet(cfg.MvmName) && utils.IsSet(cfg.MvmNamespace) {
			return nil, fmt.Errorf("required: --name")
		}
	}

	list, err := m.c.List(cfg.MvmName, cfg.MvmNamespace)
	if err != nil {
		return nil, err
	}

	if utils.IsSet(cfg.MvmName) && utils.IsSet(cfg.MvmNamespace) && !cfg.DeleteAll {
		if len(list.Microvm) > 1 {
			log.Printf("%d MicroVMs found under %s/%s:\n", len(list.Microvm), cfg.MvmNamespace, cfg.MvmName)

			var found []string

			for _, mvm := range list.Microvm {
				found = append(found, *mvm.Spec.Uid)
			}

			return found, nil
		}
	}

	for _, mvm := range list.Microvm {
		_, err := m.c.Delete(*mvm.Spec.Uid)
		if err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

func newMicroVM(name, namespace, sshPath string) (*types.MicroVMSpec, error) {
	mvm := defaults.BaseMicroVM()

	metaData, err := data.CreateMetadata(name, namespace)
	if err != nil {
		return nil, err
	}

	userData, err := data.CreateUserData(name, sshPath)
	if err != nil {
		return nil, err
	}

	mvm.Id = name
	mvm.Namespace = namespace
	mvm.Metadata = map[string]string{
		"meta-data": metaData,
		"user-data": userData,
	}

	return mvm, nil
}
