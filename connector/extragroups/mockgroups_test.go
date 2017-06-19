package extragroups

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/coreos/dex/connector"
)

func TestNewMockGroups(t *testing.T) {
	tests := []struct {
		conf     json.RawMessage
		identity connector.Identity
		groups   []string
	}{
		{
			json.RawMessage(`{}`),
			connector.Identity{},
			[]string{},
		},
		{
			json.RawMessage(`{"groups": ["one"]}`),
			connector.Identity{},
			[]string{"one"},
		},
		{
			json.RawMessage(`{"groups": ["one", "two"]}`),
			connector.Identity{},
			[]string{"one", "two"},
		},
		{
			json.RawMessage(`{"groups": []}`),
			connector.Identity{},
			[]string{},
		},
	}

	for _, test := range tests {
		f, err := newMockGroups(test.conf)
		if err != nil {
			t.Errorf("newMockGroups(%s) got error %s expected nil", test.conf, err)
		}
		got, err := f(test.identity)
		if err != nil {
			t.Errorf("mockGroup(%+v) got error %s expected nil", test.identity, err)
		}
		if !reflect.DeepEqual(got, test.groups) {
			t.Errorf("mockGroup(%+v) got %v expected %v", test.identity, got, test.groups)
		}
	}
}
