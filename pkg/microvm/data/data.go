package data

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/weaveworks/flintlock/client/cloudinit/instance"
	"github.com/weaveworks/flintlock/client/cloudinit/userdata"
	"gopkg.in/yaml.v2"

	"github.com/warehouse-13/hammertime/pkg/utils"
)

func CreateUserData(name, sshPath string) (string, error) {
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

func CreateMetadata(name, ns string) (string, error) {
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
