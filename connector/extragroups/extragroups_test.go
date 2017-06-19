package extragroups

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/dex/connector"
	"github.com/coreos/dex/connector/config"
)

func TestAddGroups(t *testing.T) {
	tests := []struct {
		eg   extragroups
		base connector.Identity
		res  connector.Identity
	}{
		{
			extragroups{[]groupSource{groupSource{prefix: "", f: newMockGroupClosure([]string{"one"})}}},
			connector.Identity{},
			connector.Identity{Groups: []string{"one"}},
		},
		{
			extragroups{[]groupSource{groupSource{prefix: "", f: newMockGroupClosure([]string{})}}},
			connector.Identity{},
			connector.Identity{},
		},
		{
			extragroups{[]groupSource{groupSource{prefix: "", f: newMockGroupClosure([]string{"one"})}}},
			connector.Identity{Groups: []string{"existing"}},
			connector.Identity{Groups: []string{"existing", "one"}},
		},
		{
			extragroups{[]groupSource{
				groupSource{prefix: "", f: newMockGroupClosure([]string{"one"})},
				groupSource{prefix: "", f: newMockGroupClosure([]string{"two"})},
			}},
			connector.Identity{Groups: []string{}},
			connector.Identity{Groups: []string{"one", "two"}},
		},
		{
			extragroups{[]groupSource{
				groupSource{prefix: "", f: newMockGroupClosure([]string{"one"})},
				groupSource{prefix: "", f: newMockGroupClosure([]string{"two"})},
			}},
			connector.Identity{Groups: []string{"existing"}},
			connector.Identity{Groups: []string{"existing", "one", "two"}},
		},
		{
			extragroups{[]groupSource{
				groupSource{prefix: "", f: newMockGroupClosure([]string{"one"})},
				groupSource{prefix: "", f: newMockGroupClosure([]string{"one"})},
			}},
			connector.Identity{Groups: []string{}},
			connector.Identity{Groups: []string{"one"}},
		},
		{
			extragroups{[]groupSource{
				groupSource{prefix: "", f: newMockGroupClosure([]string{"one"})},
				groupSource{prefix: "", f: newMockGroupClosure([]string{"two"})},
				groupSource{prefix: "", f: newMockGroupClosure([]string{"one", "three"})},
			}},
			connector.Identity{Groups: []string{}},
			connector.Identity{Groups: []string{"one", "two", "three"}},
		},
		{
			extragroups{[]groupSource{
				groupSource{f: newMockGroupClosure([]string{"one"})},
			}},
			connector.Identity{Groups: []string{}},
			connector.Identity{Groups: []string{"one"}},
		},
		{
			extragroups{[]groupSource{
				groupSource{prefix: "a-", f: newMockGroupClosure([]string{"one"})},
				groupSource{prefix: "b-", f: newMockGroupClosure([]string{"one"})},
				groupSource{prefix: "a-", f: newMockGroupClosure([]string{"one"})},
			}},
			connector.Identity{Groups: []string{}},
			connector.Identity{Groups: []string{"a-one", "b-one"}},
		},
		{
			extragroups{[]groupSource{
				groupSource{prefix: "a-", f: newMockGroupClosure([]string{"one"})},
				groupSource{prefix: "b-", f: newMockGroupClosure([]string{"one"})},
			}},
			connector.Identity{Groups: []string{}},
			connector.Identity{Groups: []string{"a-one", "b-one"}},
		},
	}

	for i, test := range tests {
		got, err := test.eg.addGroups(test.base)
		if err != nil {
			t.Errorf("addGroups(%d) got error: %s expected nil", i, err)
		}
		if !reflect.DeepEqual(got, test.res) {
			t.Errorf("addGroups(%d) got: %v expected: %v", i, got, test.res)
		}
	}

}

func TestOpen(t *testing.T) {
	rawConfig := []byte(`
connector:
  name: mock
  type: mockCallback
  id: mock
groupsources:
  - type: mock
    prefix: v1-
    config:
      groups:
        - one
        - two
  - type: mock
    prefix: v2-
    config:
      groups:
        - one
        - two
`)

	expected := []string{"v1-one", "v1-two", "v2-one", "v2-two"}
	var c Config
	if err := yaml.Unmarshal(rawConfig, &c); err != nil {
		t.Fatalf("Failed to decode %s : Error: %s", rawConfig, err)
	}
	conn, err := c.Open(logrus.New())
	if err != nil {
		t.Fatalf("Error opening config %s got: %s, expected no error", c, err)
	}
	conn, ok := conn.(connector.CallbackConnector)
	if !ok {
		t.Fatalf("Unexpected interface: %#v; expected interface connector.CallbackConnector", conn)
	}
	cbgroups, ok := conn.(*callbackgroups)
	if !ok {
		t.Fatalf("Unexpected connection type: %#v; expected *callbackgroups", conn)
	}
	i, err := cbgroups.addGroups(connector.Identity{})
	if err != nil {
		t.Fatalf("Unexpected error from addGroups(), got: %s", err)
	}
	if !reflect.DeepEqual(expected, i.Groups) {
		t.Fatalf("addGroups got: %v expected: %v", i.Groups, expected)
	}
}

func TestOpenFailure(t *testing.T) {
	// mockPassword requires a username and password this should result in a failed call
	// to connector.OpenConnector()
	c := Config{
		config.ConnectorConfig{Type: "mockPassword", Name: "foo", ID: "bar", Config: json.RawMessage("{}")},
		[]GroupSource{},
	}
	_, err := c.Open(logrus.New())
	if err == nil {
		t.Fatalf("no error opening config %s got no error, expected error", c)
	}
}
