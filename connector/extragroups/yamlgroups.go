package extragroups

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ghodss/yaml"

	"io"

	"io/ioutil"

	"github.com/coreos/dex/connector"
)

type groupmap map[string][]string

// newYamlGroup opens a "csv" type groupFunc.
func newYamlGroup(conf json.RawMessage) (groupsFunc, error) {
	config := new(struct {
		FileName string `json:"filename"`
	})
	err := json.Unmarshal(conf, config)
	if err != nil {
		return nil, fmt.Errorf("unable to parse yaml group config: %v", err)
	}
	if len(config.FileName) == 0 {
		return nil, fmt.Errorf("invalid yaml config: must specify a filename")
	}
	f, err := os.Open(config.FileName)
	if err != nil {
		return nil, fmt.Errorf("cannot open group file: %s", err)
	}
	groups, err := yamlLoadGroups(f)
	if err != nil {
		return nil, fmt.Errorf("error loading groups: %s", err)
	}
	return newYamlGroupClosure(groups), nil
}

func yamlLoadGroups(src io.ReadCloser) (groupmap, error) {
	defer src.Close()
	var groups groupmap
	data, err := ioutil.ReadAll(src)
	if err != nil {
		return groups, fmt.Errorf("error read group file: %s", err)
	}
	err = yaml.Unmarshal(data, &groups)
	if err != nil {
		return nil, fmt.Errorf("error decoding group file: %s", err)
	}
	if groups == nil {
		groups = groupmap{}
	}
	return groups, nil
}

func newYamlGroupClosure(groups groupmap) groupsFunc {
	return func(i connector.Identity) ([]string, error) {
		if g, ok := groups[i.Email]; ok {
			return g, nil
		}
		return []string{}, nil
	}
}
