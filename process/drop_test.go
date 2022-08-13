package process

var dropTests = []struct {
	name string
	proc Drop
	test [][]byte
	err  error
}{
	{
		"drop",
		Drop{},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"baz"}`),
			[]byte(`{"foo":"qux"}`),
		},
		nil,
	},
}
