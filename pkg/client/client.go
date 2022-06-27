package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks/flintlock/api/types"
	"github.com/weaveworks/flintlock/client/cloudinit/instance"
	"github.com/weaveworks/flintlock/client/cloudinit/userdata"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/yaml.v2"

	"github.com/warehouse-13/hammertime/pkg/utils"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakeclient/ github.com/weaveworks/flintlock/api/services/microvm/v1alpha1.MicroVMClient

// Client is a wrapper around a v1alpha1.MicroVMClient.
type Client struct {
	// TODO we can probably inline this #23
	flClient v1alpha1.MicroVMClient
}

// New returns a new flintlock Client.
func New(flClient v1alpha1.MicroVMClient) Client {
	return Client{
		flClient: flClient,
	}
}

// Create creates a new Microvm with the MicroVMClient.
func (c *Client) Create(name, ns, jsonSpec, sshPath string) (*v1alpha1.CreateMicroVMResponse, error) {
	var (
		mvm *types.MicroVMSpec
		err error
	)

	if utils.IsSet(jsonSpec) {
		mvm, err = loadSpecFromFile(jsonSpec)
		if err != nil {
			return nil, err
		}
	} else {
		mvm, err = defaultMicroVM(name, ns, sshPath)
		if err != nil {
			return nil, err
		}
	}

	createReq := v1alpha1.CreateMicroVMRequest{
		Microvm: mvm,
	}

	resp, err := c.flClient.CreateMicroVM(context.Background(), &createReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Get fetches a Microvm with the MicroVMClient by the given ID.
// TODO: add tests as part of #23.
func (c *Client) Get(uid string) (*v1alpha1.GetMicroVMResponse, error) {
	getReq := v1alpha1.GetMicroVMRequest{
		Uid: uid,
	}

	resp, err := c.flClient.GetMicroVM(context.Background(), &getReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// List fetches Microvms filtered by name and namespace.
// TODO: add tests as part of #23.
func (c *Client) List(name, ns string) (*v1alpha1.ListMicroVMsResponse, error) {
	listReq := v1alpha1.ListMicroVMsRequest{
		Namespace: ns,
		Name:      utils.PointyString(name),
	}

	resp, err := c.flClient.ListMicroVMs(context.Background(), &listReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Delete deletes a Microvm by the given id.
// TODO: add tests as part of #23.
func (c *Client) Delete(uid string) (*emptypb.Empty, error) {
	delReq := v1alpha1.DeleteMicroVMRequest{
		Uid: uid,
	}

	resp, err := c.flClient.DeleteMicroVM(context.Background(), &delReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func loadSpecFromFile(file string) (*types.MicroVMSpec, error) {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var spec *types.MicroVMSpec
	if err := json.Unmarshal(dat, &spec); err != nil {
		return nil, err
	}

	return spec, nil
}

// TODO: we can probably refactor this spec generation as part of #23.
func defaultMicroVM(name, namespace, sshPath string) (*types.MicroVMSpec, error) {
	var (
		kernelImage = "ghcr.io/weaveworks/flintlock-kernel:5.10.77"
		cloudImage  = "ghcr.io/weaveworks/capmvm-kubernetes:1.21.8"
	)

	metaData, err := createMetadata(name, namespace)
	if err != nil {
		return nil, err
	}

	userData, err := createUserData(name, sshPath)
	if err != nil {
		return nil, err
	}

	return &types.MicroVMSpec{
		Id:         name,
		Namespace:  namespace,
		Vcpu:       2,    //nolint: gomnd // we don't care
		MemoryInMb: 2048, //nolint: gomnd // we don't care
		Kernel: &types.Kernel{
			Image:            kernelImage,
			Filename:         utils.PointyString("boot/vmlinux"),
			AddNetworkConfig: true,
		},
		RootVolume: &types.Volume{
			Id:         "root",
			IsReadOnly: false,
			MountPoint: "/",
			Source: &types.VolumeSource{
				ContainerSource: utils.PointyString(cloudImage),
			},
		},
		Interfaces: []*types.NetworkInterface{
			{
				DeviceId: "eth1",
				Type:     0,
			},
		},
		Metadata: map[string]string{
			"meta-data": metaData,
			"user-data": userData,
		},
	}, nil
}

func createUserData(name, sshPath string) (string, error) {
	defaultUser := userdata.User{
		Name: "root",
	}

	if utils.IsSet(sshPath) {
		sshKey, err := getKeyFromPath(sshPath)
		if err != nil {
			return "", err
		}

		defaultUser.SSHAuthorizedKeys = []string{
			sshKey,
		}
	}

	// TODO: remove the boot command temporary fix after image-builder #6
	userData := &userdata.UserData{
		HostName: name,
		Users: []userdata.User{
			defaultUser,
		},
		FinalMessage: "The Liquid Metal booted system is good to go after $UPTIME seconds",
		BootCommands: []string{
			"ln -sf /run/systemd/resolve/stub-resolv.conf /etc/resolv.conf",
		},
	}

	data, err := yaml.Marshal(userData)
	if err != nil {
		return "", fmt.Errorf("marshalling bootstrap data: %w", err)
	}

	dataWithHeader := append([]byte("#cloud-config\n"), data...)

	return base64.StdEncoding.EncodeToString(dataWithHeader), nil
}

func createMetadata(name, ns string) (string, error) {
	metadata := instance.New(
		instance.WithInstanceID(fmt.Sprintf("%s/%s", ns, name)),
		instance.WithLocalHostname(name),
		instance.WithPlatform("liquid_metal"),
	)

	userMeta, err := yaml.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("unable to marshal metadata: %w", err)
	}

	return base64.StdEncoding.EncodeToString(userMeta), nil
}

func getKeyFromPath(path string) (string, error) {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(key), nil
}
