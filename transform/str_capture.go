package transform

func captureGetBytesMatch(match [][]byte) []byte {
	if len(match) > 1 {
		return match[len(match)-1]
	}

	return nil
}

func captureGetStringMatch(match []string) string {
	if len(match) > 1 {
		return match[len(match)-1]
	}

	return ""
}
