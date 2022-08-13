package process

var countTests = []struct {
	name     string
	proc     Count
	test     [][]byte
	expected []byte
	err      error
}{
	{
		"count",
		Count{},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"baz"}`),
			[]byte(`{"foo":"qux"}`),
		},
		[]byte(`{"count":3}`),
		nil,
	},
}
