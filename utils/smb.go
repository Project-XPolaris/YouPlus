package utils

func GetSmbBoolText(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
