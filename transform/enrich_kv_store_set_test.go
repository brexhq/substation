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

// func benchmarkFormatFromBase64(b *testing.B, tf *formatFromBase64, data []byte) {
// 	ctx := context.TODO()
// 	for i := 0; i < b.N; i++ {
// 		msg := message.New().SetData(data)
// 		_, _ = tf.Transform(ctx, msg)
// 	}
// }

// func BenchmarkFormatFromBase64(b *testing.B) {
// 	for _, test := range formatFromBase64Tests {
// 		tf, err := newFormatFromBase64(context.TODO(), test.cfg)
// 		if err != nil {
// 			b.Fatal(err)
// 		}

// 		b.Run(test.name,
// 			func(b *testing.B) {
// 				benchmarkFormatFromBase64(b, tf, test.test)
// 			},
// 		)
// 	}
// }
