package utils

// PointyString returns the given string's reference.
func PointyString(v string) *string {
	return &v
}

// IsSet returns true if the value of flag is not empty.
func IsSet(flag string) bool {
	return flag != ""
}
