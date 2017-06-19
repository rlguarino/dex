package extragroups

import "testing"
import "github.com/coreos/dex/connector"
import "reflect"
import "io"
import "io/ioutil"
import "strings"

func TestYamlGroupClosure(t *testing.T) {
	tests := []struct {
		gm       groupmap
		i        connector.Identity
		expected []string
	}{
		{
			groupmap{
				"foo@example.com": []string{"one"},
			},
			connector.Identity{Email: "foo@example.com"},
			[]string{"one"},
		},
		{
			groupmap{
				"foo@example.com": []string{"one"},
			},
			connector.Identity{Email: "bar@example.com"},
			[]string{},
		},
	}

	for _, test := range tests {
		f := newYamlGroupClosure(test.gm)
		res, err := f(test.i)
		if err != nil {
			t.Errorf("yamlGroupClosure() with %+v got error: %v expected no error", test.gm, err)
		}
		if !reflect.DeepEqual(res, test.expected) {
			t.Errorf("yamlGroupClosure() with %+v expected: %v got %v", test.gm, test.expected, res)
		}
	}
}

func TestYamlLoadGroups(t *testing.T) {
	tests := []struct {
		src      io.ReadCloser
		expected groupmap
	}{
		{
			ioutil.NopCloser(strings.NewReader(``)),
			groupmap{},
		},
		{
			ioutil.NopCloser(strings.NewReader(`
foo@example.com:
  - one
  - two`)),
			groupmap{
				"foo@example.com": []string{"one", "two"},
			},
		},
		{
			ioutil.NopCloser(strings.NewReader(`
foo@example.com:
  - one
bar@example.com:
  - two`)),
			groupmap{
				"foo@example.com": []string{"one"},
				"bar@example.com": []string{"two"},
			},
		},
	}

	for _, test := range tests {
		groups, err := yamlLoadGroups(test.src)
		if err != nil {
			t.Errorf("yamlLoadGroups() returned an error: %+v, expected no error", err)
		}
		if !reflect.DeepEqual(groups, test.expected) {
			t.Errorf("yamlLoadGroups() got: %#v, expected: %#v", groups, test.expected)
		}
	}
}
