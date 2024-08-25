package aws

import "strings"

func ParseRegion(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) < 4 {
		return ""
	}

	return parts[3]
}
