package process

var deleteTests = []struct {
	name     string
	proc     Delete
	test     []byte
	expected []byte
	err      error
}{
	{
		"string",
		Delete{
			InputKey: "baz",
		},
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON",
		Delete{
			InputKey: "baz",
		},
		[]byte(`{"foo":"bar","baz":{"qux":"quux"}}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
}
