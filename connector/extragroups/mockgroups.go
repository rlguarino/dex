package extragroups

import (
	"encoding/json"
	"fmt"

	"github.com/coreos/dex/connector"
)

// newMockGroups opens a "mock" type groupFunc.
func newMockGroups(conf json.RawMessage) (groupsFunc, error) {
	config := struct {
		Groups []string `json:"groups"`
	}{}
	err := json.Unmarshal(conf, &config)
	if err != nil {
		return nil, fmt.Errorf("unable to parse mock group config: %v", err)
	}
	if config.Groups == nil {
		config.Groups = []string{}
	}
	return newMockGroupClosure(config.Groups), nil
}

func newMockGroupClosure(groups []string) groupsFunc {
	return func(i connector.Identity) ([]string, error) {
		return groups, nil
	}
}
