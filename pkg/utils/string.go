package utils

func PointyString(v string) *string {
	return &v
}

func IsSet(flag string) bool {
	return flag != ""
}
