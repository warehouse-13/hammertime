package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"
	"github.com/weaveworks/flintlock/api/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/utils"
)

const (
	defaultDialTarget   = "127.0.0.1"
	defaultPort         = "9090"
	defaultMvmName      = "mvm0"
	defaultMvmNamespace = "ns0"
)

func main() { //nolint // we are refactoring this file
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
					conn, err := grpc.Dial(
						fmt.Sprintf("%s:%s", dialTarget, port),
						grpc.WithTransportCredentials(insecure.NewCredentials()),
						grpc.WithBlock(),
					)
					if err != nil {
						return err
					}
					defer conn.Close()

					client := client.New(v1alpha1.NewMicroVMClient(conn))

					res, err := client.Create(mvmName, mvmNamespace, jsonSpec, sshKeyPath)
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
					conn, err := grpc.Dial(
						fmt.Sprintf("%s:%s", dialTarget, port),
						grpc.WithTransportCredentials(insecure.NewCredentials()),
						grpc.WithBlock(),
					)
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
						fmt.Printf("%d MicroVMs found under %s/%s:\n%s",
							len(res.Microvm),
							mvmNamespace,
							mvmName,
							strings.Join(uuids, "\n"))

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
					conn, err := grpc.Dial(
						fmt.Sprintf("%s:%s", dialTarget, port),
						grpc.WithTransportCredentials(insecure.NewCredentials()),
						grpc.WithBlock(),
					)
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
					conn, err := grpc.Dial(
						fmt.Sprintf("%s:%s", dialTarget, port),
						grpc.WithTransportCredentials(insecure.NewCredentials()),
						grpc.WithBlock(),
					)
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
							return fmt.Errorf("required: --namespace")
						}
					} else {
						if isSet(mvmName) && !isSet(mvmNamespace) {
							return fmt.Errorf("required: --namespace")
						}
						if !isSet(mvmName) && isSet(mvmNamespace) {
							return fmt.Errorf("required: --name")
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

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func prettyPrint(response interface{}) error {
	resJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", string(resJSON))

	return nil
}

func isSet(flag string) bool {
	return flag != ""
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

func processFile(file string) (string, string, string, error) {
	var uid, name, namespace string

	spec, err := loadSpecFromFile(file)
	if err != nil {
		return "", "", "", err
	}

	if spec.Uid == nil && (!isSet(spec.Id) && !isSet(spec.Namespace)) {
		return "", "", "", fmt.Errorf("required: uuid or name/namespace")
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
