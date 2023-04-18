package sink

import "testing"

var filePathTests = []struct {
	name     string
	test     filePath
	expected string
}{
	{
		"prefix_suffix",
		filePath{
			Prefix: "a",
			Suffix: "z",
		},
		"a/z",
	},
	{
		"prefix",
		filePath{
			Prefix: "a",
		},
		"a",
	},
	{
		"suffix",
		filePath{
			Suffix: "z",
		},
		"z",
	},
	{
		"prefixkey_suffixkey",
		filePath{
			PrefixKey: "a",
			SuffixKey: "z",
		},
		"${PATH_PREFIX}/${PATH_SUFFIX}",
	},
	{
		"prefix_prefixkey",
		filePath{
			Prefix:    "a",
			PrefixKey: "a",
		},
		"${PATH_PREFIX}",
	},
	{
		"suffix_suffixkey",
		filePath{
			Suffix:    "z",
			SuffixKey: "z",
		},
		"${PATH_SUFFIX}",
	},
}

func TestFilePath(t *testing.T) {
	for _, test := range filePathTests {
		if test.test.New() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, test.test.New())
		}
	}
}
