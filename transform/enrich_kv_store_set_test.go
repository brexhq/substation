package transform

import (
	"reflect"
	"testing"
)

var _ Transformer = &enrichKVStoreSet{}

var kvStoreSetTruncateTests = []struct {
	name     string
	test     []byte
	expected int64
}{
	{
		"unix millisecond",
		[]byte("1696482368492"),
		1696482368,
	},
	{
		"unix nanosecond",
		[]byte("1696482368492290"),
		1696482368,
	},
}

func TestKVStoreSetTruncate(t *testing.T) {
	for _, test := range kvStoreSetTruncateTests {
		t.Run(test.name, func(t *testing.T) {
			tf := &enrichKVStoreSet{}

			tmp := bytesToValue(test.test)
			result := tf.truncateTTL(tmp)

			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}
