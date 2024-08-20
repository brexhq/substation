package base64

import (
	"encoding/base64"
	"fmt"
)

// Decode is a convenience wrapper for base64 decoding bytes.
func Decode(b []byte) ([]byte, error) {
	decode := make([]byte, base64.StdEncoding.DecodedLen(len(b)))
	n, err := base64.StdEncoding.Decode(decode, b)
	if err != nil {
		return nil, fmt.Errorf("decode: %v", err)
	}

	return decode[:n], nil
}

// Encode is a convenience wrapper for base64 encoding bytes.
func Encode(b []byte) []byte {
	encode := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(encode, b)

	return encode
}
