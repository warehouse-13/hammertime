package utils

import (
	"encoding/json"
	"fmt"
)

// PrettyPrint will render the server's response nicely in JSON.
func PrettyPrint(response interface{}) error {
	resJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", string(resJSON))

	return nil
}
