package json

import (
	"encoding/json"
	"reflect"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/brexhq/substation/internal/errors"
)

// JSONSetRawInvalid is used when SetRaw receives an invalid input
const JSONSetRawInvalid = errors.Error("JSONSetRawInvalid")

// JSONInvalidData is used when JSON functions return invalid JSON
const JSONInvalidData = errors.Error("JSONInvalidData")

// Types maps gjson.Type to strings
var Types = map[gjson.Type]string{
	0: "Null",
	1: "Boolean", // False
	2: "Number",
	3: "String",
	4: "Boolean", // True
	5: "JSON",
}

// Result wraps gjson.Result.
type Result = gjson.Result

// Delete wraps sjson.DeleteBytes.
func Delete(json []byte, key string) ([]byte, error) {
	return sjson.DeleteBytes(json, key)
}

// Get wraps gjson.GetBytes.
func Get(json []byte, key string) Result {
	return gjson.GetBytes(json, key)
}

// Set wraps sjson.SetBytes.
func Set(json []byte, key string, value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case Result:
		return sjson.SetBytes(json, key, v.Value())
	default:
		return sjson.SetBytes(json, key, v)
	}
}

// SetRaw wraps sjson.SetRawBytes.
func SetRaw(json []byte, key string, value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return sjson.SetRawBytes(json, key, v)
	case string:
		return sjson.SetRawBytes(json, key, []byte(v))
	case Result:
		return sjson.SetRawBytes(json, key, []byte(v.String()))
	default:
		return nil, JSONSetRawInvalid
	}
}

// Unmarshal wraps json.Unmarshal.
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Valid wraps json.Valid.
func Valid(data []byte) bool {
	return json.Valid(data)
}

// DeepEquals performs a deep equals comparison between two byte arrays.
func DeepEquals(s1, s2 []byte) (bool, error) {
	var j1, j2 interface{}

	if err := Unmarshal(s1, &j1); err != nil {
		return false, err
	}

	if err := Unmarshal(s2, &j2); err != nil {
		return false, err
	}

	return reflect.DeepEqual(j2, j1), nil
}
