package dialler

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// New process the dial config and returns a grpc.ClientConn. The caller is
// responsible for closing the connection.
func New(address, basicAuthToken string, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	// TODO this needs to be tidied up when adding TLS #47
	dialOpts := opts

	dialOpts = append(dialOpts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if basicAuthToken != "" {
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(
			basic(basicAuthToken),
		))
	}

	return grpc.Dial(
		address,
		dialOpts...,
	)
}
