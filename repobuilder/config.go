/*
Configuration

The RepositoryConfig object provides some basic metadata used to
generate repositories in addition to information about every
repository.
*/
package repobuilder

import (
	"fmt"
	"io/ioutil"

	"github.com/tychoish/grip"
	"gopkg.in/yaml.v2"
)

// RepositoryConfig provides an interface and schema for the
// repository configuration file. These files contain some basic
// global configuration, and a list of repositories, controled by the
// RepositoryDefinition type.
type RepositoryConfig struct {
	Mirrors   map[string]string       `bson:"mirrors" json:"mirrors" yaml:"mirrors"`
	Templates map[string]string       `bson:"templates" json:"templates" yaml:"templates"`
	Repos     []*RepositoryDefinition `bson:"repos" json:"repos" yaml:"repos"`

	fileName         string
	definitionLookup map[string]map[string]*RepositoryDefinition
}

// RepoType defines type of repositories.
type RepoType string

const (
	// RPM is a constant to refer to RPM repositories.
	RPM RepoType = "rpm"

	// DEB is a constant to refer to DEB repositories.
	DEB = "deb"
)

// RepositoryDefinition objects
type RepositoryDefinition struct {
	Name    string   `bson:"name" json:"name" yaml:"name"`
	Type    RepoType `bson:"type" json:tu"type" yaml:"type"`
	Bucket  string   `bson:"bucket" json:"bucket" yaml:"bucket"`
	Repos   []string `bson:"repos" json:"repos" yaml:"repos"`
	Edition string   `bson:"edition" json:"edition" yaml:"edition"`
}

// NewRepositoryConfig produces a pointer to an initialized
// RepositoryConfig object.
func NewRepositoryConfig() *RepositoryConfig {
	return &RepositoryConfig{
		Mirrors:          make(map[string]string),
		Templates:        make(map[string]string),
		definitionLookup: make(map[string]map[string]*RepositoryDefinition),
	}
}

// GetConfig takes the name of a file and returns a pointer to
// RepositoryConfig object. If the object is invalid or currupt in
// some way, the method returns a nil RepositoryConfig and an error.
func GetConfig(fileName string) (*RepositoryConfig, error) {
	c := NewRepositoryConfig()

	err := c.read(fileName)
	if err != nil {
		return nil, err
	}

	err = c.processRepos()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *RepositoryConfig) read(fileName string) error {
	c.fileName = fileName

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		grip.Infof("could not read file %v", fileName)
		return err
	}

	return yaml.Unmarshal(data, c)
}

func (c *RepositoryConfig) processRepos() error {
	catcher := grip.NewCatcher()

	for idx, dfn := range c.Repos {
		// do some basic validation that the type value is correct.
		if dfn.Type != DEB && dfn.Type != RPM {
			catcher.Add(fmt.Errorf("%s is not a valid repo type", dfn.Type))
		}

		// build the definitionLookup map
		if _, ok := c.definitionLookup[dfn.Edition]; !ok {
			c.definitionLookup[dfn.Edition] = make(map[string]*RepositoryDefinition)
		}

		// this lets us detect if there are duplicate
		// repository/edition pairs.
		if _, ok := c.definitionLookup[dfn.Edition][dfn.Name]; ok {
			catcher.Add(fmt.Errorf("the %s.%s already exists as repo #%d",
				dfn.Edition, dfn.Name, idx))
			continue
		}

		c.definitionLookup[dfn.Edition][dfn.Name] = dfn
	}

	return catcher.Resolve()
}

// GetRepoistoryDefinition takes the name of as repository and an edition,
// return a repository configuration. The second value is true when
// the requested edition+name exists, and false otherwise. When the
// requested edition+name does not exist, the value is nil.
func (c *RepositoryConfig) GetRepositoryDefinition(name, edition string) (*RepositoryDefinition, bool) {
	e, ok := c.definitionLookup[edition]
	if !ok {
		return nil, false
	}

	dfn, ok := e[name]
	if !ok {
		return nil, false
	}

	return dfn, true
}