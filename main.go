package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks/flintlock/api/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	defaultDialTarget   = "127.0.0.1"
	defaultPort         = "9090"
	defaultMvmName      = "mvm0"
	defaultMvmNamespace = "ns0"
)

func main() {
	var (
		dialTarget   string
		port         string
		mvmName      string
		mvmNamespace string
	)

	app := &cli.App{
		Name:  "hammertime",
		Usage: "a basic cli client to flintlock",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "grpc-address",
				Value:       defaultDialTarget,
				Aliases:     []string{"a"},
				Usage:       "flintlock server address",
				Destination: &dialTarget,
			},
			&cli.StringFlag{
				Name:        "grpc-port",
				Value:       defaultPort,
				Aliases:     []string{"p"},
				Usage:       "flintlock server port",
				Destination: &port,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "create a new microvm",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "name",
						Value:       defaultMvmName,
						Aliases:     []string{"n"},
						Usage:       "microvm name",
						Destination: &mvmName,
					},
					&cli.StringFlag{
						Name:        "namespace",
						Value:       defaultMvmNamespace,
						Aliases:     []string{"ns"},
						Usage:       "microvm namespace",
						Destination: &mvmNamespace,
					},
				},
				Action: func(c *cli.Context) error {
					conn, err := grpc.Dial(fmt.Sprintf("%s:%s", dialTarget, port), grpc.WithInsecure(), grpc.WithBlock())
					if err != nil {
						return err
					}
					defer conn.Close()

					res, err := createMicrovm(v1alpha1.NewMicroVMClient(conn), mvmName, mvmNamespace)
					if err != nil {
						return err
					}

					return prettyPrint(res)
				},
			},
			{
				Name:  "get",
				Usage: "get an existing microvm",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "name",
						Value:       defaultMvmName,
						Aliases:     []string{"n"},
						Usage:       "microvm name",
						Destination: &mvmName,
					},
					&cli.StringFlag{
						Name:        "namespace",
						Value:       defaultMvmNamespace,
						Aliases:     []string{"ns"},
						Usage:       "microvm namespace",
						Destination: &mvmNamespace,
					},
				},
				Action: func(c *cli.Context) error {
					conn, err := grpc.Dial(fmt.Sprintf("%s:%s", dialTarget, port), grpc.WithInsecure(), grpc.WithBlock())
					if err != nil {
						return err
					}
					defer conn.Close()

					res, err := getMicrovm(v1alpha1.NewMicroVMClient(conn), mvmName, mvmNamespace)
					if err != nil {
						return err
					}

					return prettyPrint(res)
				},
			},
			{
				Name:  "list",
				Usage: "list all microvms in namespace",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "namespace",
						Value:       defaultMvmNamespace,
						Aliases:     []string{"ns"},
						Usage:       "microvm namespace",
						Destination: &mvmNamespace,
					},
				},
				Action: func(c *cli.Context) error {
					conn, err := grpc.Dial(fmt.Sprintf("%s:%s", dialTarget, port), grpc.WithInsecure(), grpc.WithBlock())
					if err != nil {
						return err
					}
					defer conn.Close()

					res, err := listMicrovms(v1alpha1.NewMicroVMClient(conn), mvmName, mvmNamespace)
					if err != nil {
						return err
					}

					return prettyPrint(res)
				},
			},
			{
				Name:  "delete",
				Usage: "delete an existing microvm",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "name",
						Value:       defaultMvmName,
						Aliases:     []string{"n"},
						Usage:       "microvm name",
						Destination: &mvmName,
					},
					&cli.StringFlag{
						Name:        "namespace",
						Value:       defaultMvmNamespace,
						Aliases:     []string{"ns"},
						Usage:       "microvm namespace",
						Destination: &mvmNamespace,
					},
				},
				Action: func(c *cli.Context) error {
					conn, err := grpc.Dial(fmt.Sprintf("%s:%s", dialTarget, port), grpc.WithInsecure(), grpc.WithBlock())
					if err != nil {
						return err
					}
					defer conn.Close()

					res, err := deleteMicroVM(v1alpha1.NewMicroVMClient(conn), mvmName, mvmNamespace)
					if err != nil {
						return err
					}

					return prettyPrint(res)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func prettyPrint(response interface{}) error {
	resJson, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(resJson))

	return nil
}

func createMicrovm(client v1alpha1.MicroVMClient, name, ns string) (*v1alpha1.CreateMicroVMResponse, error) {
	createReq := v1alpha1.CreateMicroVMRequest{
		Microvm: defaultMicroVM(name, ns),
	}
	resp, err := client.CreateMicroVM(context.Background(), &createReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getMicrovm(client v1alpha1.MicroVMClient, name, ns string) (*v1alpha1.GetMicroVMResponse, error) {
	getReq := v1alpha1.GetMicroVMRequest{
		Id:        name,
		Namespace: ns,
	}
	resp, err := client.GetMicroVM(context.Background(), &getReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func deleteMicroVM(client v1alpha1.MicroVMClient, name, ns string) (*emptypb.Empty, error) {
	delReq := v1alpha1.DeleteMicroVMRequest{
		Id:        name,
		Namespace: ns,
	}
	resp, err := client.DeleteMicroVM(context.Background(), &delReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func listMicrovms(client v1alpha1.MicroVMClient, name, ns string) (*v1alpha1.ListMicroVMsResponse, error) {
	listReq := v1alpha1.ListMicroVMsRequest{
		Namespace: ns,
	}
	resp, err := client.ListMicroVMs(context.Background(), &listReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func defaultMicroVM(name, namespace string) *types.MicroVMSpec {
	var (
		kernelImage = "docker.io/richardcase/ubuntu-bionic-kernel:0.0.11"
		cloudImage  = "docker.io/richardcase/ubuntu-bionic-test:cloudimage_v0.0.1"
	)

	return &types.MicroVMSpec{
		Id:         name,
		Namespace:  namespace,
		Vcpu:       2,
		MemoryInMb: 2048,
		Kernel: &types.Kernel{
			Image:            kernelImage,
			Cmdline:          "console=ttyS0 reboot=k panic=1 pci=off i8042.noaux i8042.nomux i8042.nopnp i8042.dumbkbd ds=nocloud-net;s=http://169.254.169.254/latest/",
			Filename:         pointyString("vmlinux"),
			AddNetworkConfig: true,
		},
		Initrd: &types.Initrd{
			Image:    kernelImage,
			Filename: pointyString("initrd-generic"),
		},
		RootVolume: &types.Volume{
			Id:         "root",
			IsReadOnly: true,
			MountPoint: "/",
			Source: &types.VolumeSource{
				ContainerSource: pointyString(cloudImage),
			},
		},
		Interfaces: []*types.NetworkInterface{
			{
				GuestDeviceName: "eth0",
				Type:            0,
			},
		},
		Metadata: map[string]string{
			"meta-data": "aW5zdGFuY2VfaWQ6IG5zMS9tdm0wCmxvY2FsX2hvc3RuYW1lOiBtdm0wCnBsYXRmb3JtOiBsaXF1aWRfbWV0YWwK",
			"user-data": "I2Nsb3VkLWNvbmZpZwpob3N0bmFtZTogbXZtMApmcWRuOiBtdm0wLmZydWl0Y2FzZQp1c2VyczoKICAgIC0gbmFtZTogcm9vdAogICAgICBzc2hfYXV0aG9yaXplZF9rZXlzOgogICAgICAgIC0gfAogICAgICAgICAgc3NoLWVkMjU1MTkgQUFBQUMzTnphQzFsWkRJMU5URTVBQUFBSUdzbStWSSsyVk5WWFBDRmVmbFhrQTVKY21zMzByajFGUFFjcFNTdDFrdVYgcmljaGFyZEB3ZWF2ZS53b3JrcwpkaXNhYmxlX3Jvb3Q6IGZhbHNlCnBhY2thZ2VfdXBkYXRlOiBmYWxzZQpmaW5hbF9tZXNzYWdlOiBUaGUgcmVpZ25pdGVkIGJvb3RlZCBzeXN0ZW0gaXMgZ29vZCB0byBnbyBhZnRlciAkVVBUSU1FIHNlY29uZHMKcnVuY21kOgogICAgLSBkaGNsaWVudCAtcgogICAgLSBkaGNsaWVudAo=",
		},
	}
}

func pointyString(v string) *string {
	return &v
}
