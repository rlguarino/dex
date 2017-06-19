package config

import (
	"encoding/json"

	"github.com/coreos/dex/storage"
)

// ConnectorConfig represents the configuration of a connector in the dex configuration
// file.
type ConnectorConfig struct {
	Type   string          `json:"type"`
	Name   string          `json:"name"`
	ID     string          `json:"id"`
	Config json.RawMessage `json:"config"`
}

// ToStorageConnector converts an object to storage connector type.
func ToStorageConnector(c ConnectorConfig) (storage.Connector, error) {
	return storage.Connector{
		ID:     c.ID,
		Type:   c.Type,
		Name:   c.Name,
		Config: c.Config,
	}, nil
}
