package auth

import (
	"context"
	"encoding/base64"
)

type basicAuth struct {
	token string
}

func Basic(t string) basicAuth {
	return basicAuth{token: t}
}

func (b basicAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	enc := base64.StdEncoding.EncodeToString([]byte(b.token))

	return map[string]string{
		"authorization": "Basic " + enc,
	}, nil
}

func (basicAuth) RequireTransportSecurity() bool {
	return false
}
