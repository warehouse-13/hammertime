package dialler

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// New process the dial config and returns a grpc.ClientConn. The caller is
// responsible for closing the connection.
func New(address, basicAuthToken string) (*grpc.ClientConn, error) {

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if basicAuthToken != "" {
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(
			Basic(basicAuthToken),
		))
	}

	return grpc.Dial(
		address,
		dialOpts...,
	)
}
