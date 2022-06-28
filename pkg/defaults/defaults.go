package defaults

import (
	"github.com/weaveworks/flintlock/api/types"

	"github.com/warehouse-13/hammertime/pkg/utils"
)

const (
	// DialTarget is the default address which the client will attempt to contact
	// the flintlock server on.
	DialTarget = "127.0.0.1:9090"
	// MvmName is the default name to use when creating a Microvm.
	MvmName = "mvm0"
	// MvmNamespace is the default name to use when creating a Microvm.
	MvmNamespace = "ns0"
)

const (
	// KernelImage is the default MVM kernel image.
	KernelImage = "ghcr.io/weaveworks/flintlock-kernel:5.10.77"
	// CloudImage is the default MVM cloud image.
	CloudImage = "ghcr.io/weaveworks/capmvm-kubernetes:1.21.8"
)

func BaseMicroVM() *types.MicroVMSpec {
	return &types.MicroVMSpec{
		Vcpu:       2,    //nolint: gomnd // we don't care
		MemoryInMb: 2048, //nolint: gomnd // we don't care
		Kernel: &types.Kernel{
			Image:            KernelImage,
			Filename:         utils.PointyString("boot/vmlinux"),
			AddNetworkConfig: true,
		},
		RootVolume: &types.Volume{
			Id:         "root",
			IsReadOnly: false,
			MountPoint: "/",
			Source: &types.VolumeSource{
				ContainerSource: utils.PointyString(CloudImage),
			},
		},
		Interfaces: []*types.NetworkInterface{
			{
				DeviceId: "eth1",
				Type:     0,
			},
		},
	}
}
