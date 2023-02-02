package defaults

import (
	"github.com/weaveworks-liquidmetal/flintlock/api/types"
	"k8s.io/utils/pointer"
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
	KernelImage = "ghcr.io/weaveworks-liquidmetal/kernel-bin:5.10.77"
	// ModulesImage is the default MVM kernel image.
	ModulesImage = "ghcr.io/weaveworks-liquidmetal/kernel-modules:5.10.77"
	// OSImage is the default MVM OS image.
	OSImage = "ghcr.io/weaveworks-liquidmetal/capmvm-k8s-os:1.23.5"

	kernelFilename = "boot/vmlinux"
	modulesPath    = "/lib/modules/5.10.77"
)

func BaseMicroVM() *types.MicroVMSpec {
	return &types.MicroVMSpec{
		Vcpu:       2,    //nolint: gomnd // we don't care
		MemoryInMb: 2048, //nolint: gomnd // we don't care
		Kernel: &types.Kernel{
			Image:            KernelImage,
			Filename:         pointer.String(kernelFilename),
			AddNetworkConfig: true,
		},
		RootVolume: &types.Volume{
			Id:         "root",
			IsReadOnly: false,
			Source: &types.VolumeSource{
				ContainerSource: pointer.String(OSImage),
			},
		},
		AdditionalVolumes: []*types.Volume{
			{
				Id:         "modules",
				IsReadOnly: false,
				Source: &types.VolumeSource{
					ContainerSource: pointer.String(ModulesImage),
				},
				MountPoint: pointer.String(modulesPath),
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
