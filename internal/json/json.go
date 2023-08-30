package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/brexhq/substation/internal/base64"
)

// errSetRawInvalid is returned when SetRaw receives an invalid input value.
var errSetRawInvalid = fmt.Errorf("invalid value interface")

// Types maps gjson.Type to strings.
var Types = map[gjson.Type]string{
	0: "Null",
	1: "Boolean", // False
	2: "Number",
	3: "String",
	4: "Boolean", // True
	5: "JSON",
}

var opts = &sjson.Options{
	Optimistic:     true,
	ReplaceInPlace: true,
}

type Result struct {
	gjson.Result
}

func (r Result) Bytes() []byte {
	return []byte(r.String())
}

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
	return Result{gjson.GetBytes(json, key)}
}

/*
Set inserts values into JSON and operates under these conditions (in order):

- If the value is valid JSON (bytes, string, or Result), then it is inserted using SetRaw to properly insert it as nested JSON; this avoids encoding that would otherwise create invalid JSON (e.g. `{\"foo\":\"bar\"}`)

- If the value is bytes, then it is converted to a base64 encoded string (this is the behavior of the standard library's encoding/json package)

- If the value is Result, then it is converted to the underlying GJSON Value

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
			tmp, err = sjson.SetBytesOptions(json, key, v, opts)
		} else {
			tmp, err = sjson.SetBytesOptions(json, key, base64.Encode(v), opts)
		}
	case gjson.Result:
		tmp, err = sjson.SetBytesOptions(json, key, v.Value(), opts)
	default:
		tmp, err = sjson.SetBytesOptions(json, key, v, opts)
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
		tmp, err = sjson.SetRawBytesOptions(json, key, v, opts)
	case string:
		tmp, err = sjson.SetRawBytesOptions(json, key, []byte(v), opts)
	case gjson.Result:
		tmp, err = sjson.SetRawBytesOptions(json, key, []byte(v.String()), opts)
	default:
		return nil, fmt.Errorf("set_raw key %s: %v", key, errSetRawInvalid)
	}

	if err != nil {
		return nil, fmt.Errorf("set_raw key %s: %v", key, err)
	}

	return tmp, nil
}

// Valid conditionally checks if bytes, strings, or Results are valid JSON objects.
func Valid(data interface{}) bool {
	switch v := data.(type) {
	case []byte:
		if !bytes.HasPrefix(v, []byte(`{`)) && !bytes.HasPrefix(v, []byte(`[`)) {
			return false
		}

		return json.Valid(v)
	case string:
		if !strings.HasPrefix(v, `{`) && !strings.HasPrefix(v, `[`) {
			return false
		}

		return json.Valid([]byte(v))
	// Result can have one of many underlying structs, so we need to check for multiple conditions.
	case gjson.Result:
		if v.IsObject() {
			return true
		}

		s := v.String()
		if !strings.HasPrefix(s, `{`) && !strings.HasPrefix(s, `[`) {
			return false
		}

		return json.Valid([]byte(s))
	default:
		return false
	}
}
