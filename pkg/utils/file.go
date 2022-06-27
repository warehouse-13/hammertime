package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/weaveworks/flintlock/api/types"
)

// ProcessFile will open the given file and process the JSON into a MicroVMSpec.
func ProcessFile(file string) (string, string, string, error) {
	var uid, name, namespace string

	spec, err := LoadSpecFromFile(file)
	if err != nil {
		return "", "", "", err
	}

	if spec.Uid == nil && (!IsSet(spec.Id) && !IsSet(spec.Namespace)) {
		return "", "", "", fmt.Errorf("required: uuid or name/namespace")
	}

	if spec.Uid != nil {
		uid = *spec.Uid
	}

	name = spec.Id
	namespace = spec.Namespace

	return uid, name, namespace, nil
}

// LoadSpecFromFile reads the given json file and produces a MicroVMSpec.
func LoadSpecFromFile(file string) (*types.MicroVMSpec, error) {
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
