package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/Callisto13/hammertime/pkg/utils"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks/flintlock/api/types"
	"github.com/weaveworks/flintlock/client/cloudinit/instance"
	"github.com/weaveworks/flintlock/client/cloudinit/userdata"
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
		mvmUID       string
		mvmNamespace string
		sshKeyPath   string
		jsonSpec     string
		state        bool
		deleteAll    bool
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
					&cli.StringFlag{
						Name:        "public-key-path",
						Aliases:     []string{"k"},
						Usage:       "path to file containing public SSH key to be added to root user",
						Destination: &sshKeyPath,
					},
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Usage:       "path to json file containing full flintlock spec. will override other flags",
						Destination: &jsonSpec,
					},
				},
				Action: func(c *cli.Context) error {
					conn, err := grpc.Dial(fmt.Sprintf("%s:%s", dialTarget, port), grpc.WithInsecure(), grpc.WithBlock())
					if err != nil {
						return err
					}
					defer conn.Close()

					res, err := createMicrovm(v1alpha1.NewMicroVMClient(conn), mvmName, mvmNamespace, sshKeyPath, jsonSpec)
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
						Name:        "id",
						Aliases:     []string{"i"},
						Usage:       "microvm uuid",
						Destination: &mvmUID,
					},
					&cli.BoolFlag{
						Name:        "state",
						Value:       false,
						Aliases:     []string{"s"},
						Usage:       "print just the state of the microvm",
						Destination: &state,
					},
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Usage:       "path to json file containing full flintlock spec. will override name and namespace flags",
						Destination: &jsonSpec,
					},
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

					if isSet(jsonSpec) {
						var err error
						mvmUID, mvmName, mvmNamespace, err = processFile(jsonSpec)
						if err != nil {
							return err
						}
					}

					if isSet(mvmUID) {
						res, err := getMicrovm(v1alpha1.NewMicroVMClient(conn), mvmUID)
						if err != nil {
							return err
						}

						if state {
							fmt.Println(res.Microvm.Status.State)

							return nil
						}

						return prettyPrint(res)
					}

					res, err := listMicrovms(v1alpha1.NewMicroVMClient(conn), mvmName, mvmNamespace)
					if err != nil {
						return err
					}

					if len(res.Microvm) > 1 {
						var uuids []string
						for _, microvm := range res.Microvm {
							uuids = append(uuids, *microvm.Spec.Uid)
						}
						fmt.Printf("%d MicroVMs found under %s/%s:\n%s", len(res.Microvm), mvmNamespace, mvmName, strings.Join(uuids, "\n"))

						return nil
					}

					if len(res.Microvm) == 1 {
						if state {
							fmt.Println(res.Microvm[0].Status.State)

							return nil
						}

						return prettyPrint(res.Microvm[0])
					}

					return fmt.Errorf("MicroVM %s/%s not found", mvmNamespace, mvmName)
				},
			},
			{
				Name:  "list",
				Usage: "list all microvms across all namespaces",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "namespace",
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

					res, err := listMicrovms(v1alpha1.NewMicroVMClient(conn), "", mvmNamespace)
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
						Name:        "id",
						Aliases:     []string{"i"},
						Usage:       "microvm uid",
						Destination: &mvmUID,
					},
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Usage:       "path to json file containing full flintlock spec. will override other flags",
						Destination: &jsonSpec,
					},
					&cli.BoolFlag{
						Name:        "all",
						Aliases:     []string{"a"},
						Usage:       "delete all microvms (filter with --name and --namespace)",
						Destination: &deleteAll,
					},
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "delete all microvms under this name in the given namespace",
						Destination: &mvmName,
					},
					&cli.StringFlag{
						Name:        "namespace",
						Aliases:     []string{"ns"},
						Usage:       "delete all microvms under this namespace",
						Destination: &mvmNamespace,
					},
				},
				Action: func(c *cli.Context) error {
					conn, err := grpc.Dial(fmt.Sprintf("%s:%s", dialTarget, port), grpc.WithInsecure(), grpc.WithBlock())
					if err != nil {
						return err
					}
					defer conn.Close()

					if isSet(jsonSpec) {
						var err error
						mvmUID, mvmName, mvmNamespace, err = processFile(jsonSpec)
						if err != nil {
							return err
						}
					}

					if isSet(mvmUID) {
						res, err := deleteMicroVM(v1alpha1.NewMicroVMClient(conn), mvmUID)
						if err != nil {
							return err
						}

						return prettyPrint(res)
					}

					if deleteAll {
						if isSet(mvmName) && !isSet(mvmNamespace) {
							return errors.New("required: --namespace")
						}
					} else {
						if isSet(mvmName) && !isSet(mvmNamespace) {
							return errors.New("required: --namespace")
						}
						if !isSet(mvmName) && isSet(mvmNamespace) {
							return errors.New("required: --name")
						}
					}

					list, err := listMicrovms(v1alpha1.NewMicroVMClient(conn), mvmName, mvmNamespace)
					if err != nil {
						return err
					}

					if isSet(mvmName) && isSet(mvmNamespace) && !deleteAll {
						if len(list.Microvm) > 1 {
							fmt.Printf("%d MicroVMs found under %s/%s:\n", len(list.Microvm), mvmNamespace, mvmName)
							for _, mvm := range list.Microvm {
								fmt.Println(*mvm.Spec.Uid)
							}
							return nil
						}
					}

					for _, mvm := range list.Microvm {
						res, err := deleteMicroVM(v1alpha1.NewMicroVMClient(conn), *mvm.Spec.Uid)
						if err != nil {
							return err
						}
						if err := prettyPrint(res); err != nil {
							return err
						}
					}

					return nil
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

func isSet(flag string) bool {
	return flag != ""
}

func createMicrovm(client v1alpha1.MicroVMClient, name, ns, sshPath, jsonSpec string) (*v1alpha1.CreateMicroVMResponse, error) {
	var (
		mvm *types.MicroVMSpec
		err error
	)

	if jsonSpec != "" {
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
	resp, err := client.CreateMicroVM(context.Background(), &createReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getMicrovm(client v1alpha1.MicroVMClient, uid string) (*v1alpha1.GetMicroVMResponse, error) {
	getReq := v1alpha1.GetMicroVMRequest{
		Uid: uid,
	}
	resp, err := client.GetMicroVM(context.Background(), &getReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func deleteMicroVM(client v1alpha1.MicroVMClient, uid string) (*emptypb.Empty, error) {
	delReq := v1alpha1.DeleteMicroVMRequest{
		Uid: uid,
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
		Name:      utils.PointyString(name),
	}
	resp, err := client.ListMicroVMs(context.Background(), &listReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

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
		Vcpu:       2,
		MemoryInMb: 2048,
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

	if sshPath != "" {
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

func processFile(file string) (string, string, string, error) {
	var uid, name, namespace string

	spec, err := loadSpecFromFile(file)
	if err != nil {
		return "", "", "", err
	}

	if spec.Uid == nil && (!isSet(spec.Id) && !isSet(spec.Namespace)) {
		return "", "", "", errors.New("required: uuid or name/namespace")
	}

	if spec.Uid != nil {
		uid = *spec.Uid
	}

	name = spec.Id
	namespace = spec.Namespace

	return uid, name, namespace, nil
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
