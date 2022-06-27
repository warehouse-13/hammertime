package config

type Config struct {
	// GRPCAddress is the flintlock server address.
	GRPCAddress string
	// MvmName is the name of the Microvm.
	MvmName string
	// MvmNamespace is the namespace of the Microvm.
	MvmNamespace string
	// JSONFile is the path to a file containing a Microvm Spec in json.
	JSONFile string
	// SSHKeyPath is the path to a file containing a public key. Added for
	// creating/using a Microvm with SSH access.
	SSHKeyPath string
	// State reports on only the state of a Microvm. Can only be used with `get`.
	State bool
	// DeleteAll configures all microvms to be deleted. Can only be used with `delete`.
	DeleteAll bool
	// UUID is the id of a created Microvm.
	UUID string
}
