package client

import (
	"context"

	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks/flintlock/api/types"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/warehouse-13/hammertime/pkg/utils"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakeclient/ github.com/weaveworks/flintlock/api/services/microvm/v1alpha1.MicroVMClient

// Client is a wrapper around a v1alpha1.MicroVMClient.
type Client struct {
	v1alpha1.MicroVMClient
}

// New returns a new flintlock Client.
func New(flClient v1alpha1.MicroVMClient) Client {
	return Client{
		flClient,
	}
}

// Create creates a new Microvm with the MicroVMClient.
func (c *Client) Create(mvm *types.MicroVMSpec) (*v1alpha1.CreateMicroVMResponse, error) {
	createReq := v1alpha1.CreateMicroVMRequest{
		Microvm: mvm,
	}

	return c.CreateMicroVM(context.Background(), &createReq)
}

// Get fetches a Microvm with the MicroVMClient by the given ID.
func (c *Client) Get(uid string) (*v1alpha1.GetMicroVMResponse, error) {
	getReq := v1alpha1.GetMicroVMRequest{
		Uid: uid,
	}

	return c.GetMicroVM(context.Background(), &getReq)
}

// List fetches Microvms filtered by name and namespace.
func (c *Client) List(name, ns string) (*v1alpha1.ListMicroVMsResponse, error) {
	listReq := v1alpha1.ListMicroVMsRequest{
		Namespace: ns,
		Name:      utils.PointyString(name),
	}

	return c.ListMicroVMs(context.Background(), &listReq)
}

// Delete deletes a Microvm by the given id.
func (c *Client) Delete(uid string) (*emptypb.Empty, error) {
	delReq := v1alpha1.DeleteMicroVMRequest{
		Uid: uid,
	}

	return c.DeleteMicroVM(context.Background(), &delReq)
}
