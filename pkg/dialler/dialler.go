package dialler

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// New process the dial config and returns a grpc.ClientConn. The caller is
// responsible for closing the connection.
func New(address string) (*grpc.ClientConn, error) {
	return grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}
