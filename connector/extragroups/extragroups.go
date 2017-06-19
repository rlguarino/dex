package extragroups

import (
	"context"
	"net/http"

	"fmt"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/dex/connector"
	"github.com/coreos/dex/connector/config"
	"github.com/coreos/dex/connector/github"
	"github.com/coreos/dex/connector/gitlab"
	"github.com/coreos/dex/connector/ldap"
	"github.com/coreos/dex/connector/mock"
	"github.com/coreos/dex/connector/oidc"
	"github.com/coreos/dex/connector/saml"
)

var configMap = map[string]func() connector.Factory{
	"mockCallback": func() connector.Factory { return new(mock.CallbackConfig) },
	"mockPassword": func() connector.Factory { return new(mock.PasswordConfig) },
	"ldap":         func() connector.Factory { return new(ldap.Config) },
	"github":       func() connector.Factory { return new(github.Config) },
	"gitlab":       func() connector.Factory { return new(gitlab.Config) },
	"oidc":         func() connector.Factory { return new(oidc.Config) },
	"saml":         func() connector.Factory { return new(saml.Config) },
}

// Config holds configuration options for the extragroup connector.
type Config struct {
	Connector    config.ConnectorConfig `json:"connector"`
	GroupSources []GroupSource          `json:"groupsources"`
}

// GroupSource holds generic configuration for for a group source and it's type.
type GroupSource struct {
	Type   string          `json:"type"`
	Prefix string          `json:"prefix"`
	Config json.RawMessage `json:"config"`
}

// a groupsFunc takes a connector.Identity and returns a list of groups to add to that identity.
type groupsFunc func(connector.Identity) ([]string, error)

// Open returns a connector which wraps the another connector to allow adding additional groups
// to a Identity.
func (c *Config) Open(logger logrus.FieldLogger) (connector.Connector, error) {
	var conn connector.Connector
	s, err := config.ToStorageConnector(c.Connector)
	if err != nil {
		return &conn, fmt.Errorf("failed to parse connector: %v", err)
	}
	baseconn, err := connector.OpenConnector(logger, s, configMap)
	if err != nil {
		return &conn, fmt.Errorf("failed to open connector: %v", err)
	}
	groupsrcs := []groupSource{}
	for _, gs := range c.GroupSources {
		f, err := openGroupFunc(logger, gs)
		if err != nil {
			return &conn, fmt.Errorf("failed to parse group source: %v", err)
		}
		groupsrcs = append(groupsrcs, groupSource{prefix: gs.Prefix, f: f})
	}
	switch baseconn := baseconn.(type) {
	case connector.CallbackConnector:
		return &callbackgroups{baseconnector: baseconn, extragroups: extragroups{groupSources: groupsrcs}}, nil
	case connector.SAMLConnector:
		return &samlgroups{baseconnector: baseconn, extragroups: extragroups{groupSources: groupsrcs}}, nil
	case connector.PasswordConnector:
		return &passwordgroups{baseconnector: baseconn, extragroups: extragroups{groupSources: groupsrcs}}, nil
	default:
		return &conn, fmt.Errorf("invalid connector: %v", c.Connector.Name)
	}
}

// openGroupFunc opens a group function given a group source. This function turns a
// generic GroupSource configuration into a concrete implementation.
func openGroupFunc(logger logrus.FieldLogger, gs GroupSource) (groupsFunc, error) {
	switch gs.Type {
	case "mock":
		return newMockGroups(gs.Config)
	case "yaml":
		return newYamlGroup(gs.Config)
	default:
		return nil, fmt.Errorf("No group function found: %v", gs.Type)
	}
}

type groupSource struct {
	prefix string
	f      groupsFunc
}

// extragroup is a base struct which knows how to call muptiple group functions and concatonate
// the groups return by each into one group slice.
type extragroups struct {
	groupSources []groupSource
}

// addGroups takes an identity and calls all of the group functions in sqeuence with that Itentity.
func (c *extragroups) addGroups(identity connector.Identity) (connector.Identity, error) {
	for _, src := range c.groupSources {
		newGrps, err := src.f(identity)
		if err != nil {
			return identity, err
		}
		identity.Groups = removeDuplicates(append(identity.Groups, addPrefix(src.prefix, newGrps...)...))
	}
	return identity, nil
}

func addPrefix(prefix string, groups ...string) []string {
	var res []string
	for _, g := range groups {
		res = append(res, prefix+g)
	}
	return res
}

func removeDuplicates(a []string) []string {
	if a == nil {
		return a
	}
	result := []string{}
	seen := map[string]string{}
	for _, val := range a {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = val
		}
	}
	return result
}

// extragroups wrapper which implements connector.CallbackConnector
type callbackgroups struct {
	extragroups
	baseconnector connector.CallbackConnector
}

func (c *callbackgroups) LoginURL(s connector.Scopes, callbackURL, state string) (string, error) {
	return c.baseconnector.LoginURL(s, callbackURL, state)
}

func (c *callbackgroups) HandleCallback(s connector.Scopes, r *http.Request) (identity connector.Identity, err error) {
	identity, err = c.baseconnector.HandleCallback(s, r)
	if err != nil {
		return
	}
	return c.addGroups(identity)
}

// extragroups wraper which implements connector.SAMLConnector
type samlgroups struct {
	extragroups
	baseconnector connector.SAMLConnector
}

func (c *samlgroups) POSTData(s connector.Scopes, requestID string) (sooURL, samlRequest string, err error) {
	return c.baseconnector.POSTData(s, requestID)
}

func (c *samlgroups) HandlePOST(s connector.Scopes, samlResponse, inResponseTo string) (identity connector.Identity, err error) {
	identity, err = c.baseconnector.HandlePOST(s, samlResponse, inResponseTo)
	if err != nil {
		return
	}
	return c.addGroups(identity)
}

type passwordgroups struct {
	extragroups
	baseconnector connector.PasswordConnector
}

func (c *passwordgroups) Login(ctx context.Context, s connector.Scopes, username, password string) (identity connector.Identity, validPassword bool, err error) {
	return c.baseconnector.Login(ctx, s, username, password)
}
