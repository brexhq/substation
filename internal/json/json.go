package json

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unicode/utf8"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/brexhq/substation/internal/base64"
	"github.com/brexhq/substation/internal/errors"
)

// JSONSetRawInvalid is returned when SetRaw receives an invalid input.
const JSONSetRawInvalid = errors.Error("JSONSetRawInvalid")

// JSONInvalidData is returned when JSON functions return invalid JSON.
const JSONInvalidData = errors.Error("JSONInvalidData")

// Types maps gjson.Type to strings.
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
func Delete(json []byte, key string) (tmp []byte, err error) {
	tmp, err = sjson.DeleteBytes(json, key)

	if err != nil {
		return nil, fmt.Errorf("delete key %s: %v", key, err)
	}

	return tmp, nil
}

// Get wraps gjson.GetBytes.
func Get(json []byte, key string) Result {
	return gjson.GetBytes(json, key)
}

/*
Set inserts values into JSON and operates under these conditions (in order):

- If the value is valid JSON (bytes, string, or Result), then it is inserted using SetRaw to properly insert it as nested JSON; this avoids encoding that would otherwise create invalid JSON (e.g. `{\"hello\":\"world\"}`)

-If the value is bytes, then it is converted to a base64 encoded string (this is the behavior of the standard library's encoding/json package)

- If the value is Result, then it is converted to the underlying gjson Value

- All other values are inserted as interfaces and are converted by SJSON to the proper format
*/
func Set(json []byte, key string, value interface{}) (tmp []byte, err error) {
	if Valid(value) {
		tmp, err = SetRaw(json, key, value)
		return tmp, err
	}

	switch v := value.(type) {
	case []byte:
		if utf8.Valid(v) {
			tmp, err = sjson.SetBytes(json, key, v)
		} else {
			tmp, err = sjson.SetBytes(json, key, base64.Encode(v))
		}
	case Result:
		tmp, err = sjson.SetBytes(json, key, v.Value())
	default:
		tmp, err = sjson.SetBytes(json, key, v)
	}

	if err != nil {
		return nil, fmt.Errorf("set key %s: %v", key, err)
	}

	return tmp, nil
}

// SetRaw wraps sjson.SetRawBytes and conditionally converts values to properly insert them as nested JSON.
func SetRaw(json []byte, key string, value interface{}) (tmp []byte, err error) {
	switch v := value.(type) {
	case []byte:
		tmp, err = sjson.SetRawBytes(json, key, v)
	case string:
		tmp, err = sjson.SetRawBytes(json, key, []byte(v))
	case Result:
		tmp, err = sjson.SetRawBytes(json, key, []byte(v.String()))
	default:
		return nil, fmt.Errorf("setraw key %s: %v", key, JSONSetRawInvalid)
	}

	if err != nil {
		return nil, fmt.Errorf("setraw key %s: %v", key, err)
	}

	return tmp, nil
}

// Unmarshal wraps json.Unmarshal.
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Valid wraps json.Valid and conditionally checks if bytes, strings, or Results are valid.
func Valid(data interface{}) bool {
	switch v := data.(type) {
	case []byte:
		return json.Valid(v)
	case string:
		return json.Valid([]byte(v))
	case Result:
		return json.Valid([]byte(v.String()))
	default:
		return false
	}
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
