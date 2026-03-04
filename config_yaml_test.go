package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/serega-cpp/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// The Time type is used here to show how to implement customly parsed
// types. To support them in the Yaml file  parser (gopkg.in/yaml),
// you need to implement `UnmarshalYAML` method.
//
// type Time struct {
// 	time.Time
// }

func (t *Time) UnmarshalYAML(value *yaml.Node) error {
	var buf string
	if err := value.Decode(&buf); err != nil {
		return err
	}
	return t.Set(buf)
}

// type Url struct {
// 	*url.URL
// }

func (u *Url) UnmarshalYAML(value *yaml.Node) error {
	var buf string
	if err := value.Decode(&buf); err != nil {
		return err
	}
	return u.Set(buf)
}

////////////////////////////////////////////////////////
// Config struct declaration

type Specific struct {
	Time  time.Time      `usage:"::time field"`
	Slice []string       `usage:"::slice field"`
	Map   map[string]int `usage:"::map field"`
}

type YamlConfig struct {
	Id       string `usage:"Identificator"`
	Custom   Custom
	Specific Specific
}

////////////////////////////////////////////////////////
// Sample config content

const YamlContent = "id: custom_yaml\n" +
	"custom:\n" +
	"  stdtime: 2023-02-16T12:00:00Z\n" +
	"  longtime: 2023-02-16T18:00:00\n" +
	"  shorttime: 2025-12-21\n" +
	"  link: https://google.com/search?q=golang\n" +
	"specific:\n" +
	"  time: 2020-03-23T11:15:00Z\n" +
	"  slice:\n" +
	"    - one\n" +
	"    - two\n" +
	"    - six\n" +
	"  map:\n" +
	"    one: 1\n" +
	"    two: 2\n" +
	"    six: 6\n"

///////////////////////////////////////////////////////////
// Config expected

var idYamlExpected = "custom_yaml"
var specificExpected = Specific{
	Time:  time.Date(2020, 3, 23, 11, 15, 0, 0, time.UTC),
	Slice: []string{"one", "two", "six"},
	Map: map[string]int{
		"one": 1,
		"two": 2,
		"six": 6,
	},
}

/////////////////////////////////////////////////////////
// Tests implementations

func TestConfigYaml(t *testing.T) {
	fname := createTmpFile(t, "test.yaml", YamlContent)
	t.Run("YAML", func(t *testing.T) {
		cfg, err := config.New[YamlConfig](nil).WithFile(fname,
			func(content []byte) (YamlConfig, error) {
				var cfg YamlConfig
				err := yaml.Unmarshal(content, &cfg)
				return cfg, err
			},
		).AsStruct()
		require.NoError(t, err)

		require.Equal(t, idYamlExpected, cfg.Id)
		require.Equal(t, customExpected, cfg.Custom)
		require.Equal(t, specificExpected, cfg.Specific)
	})
}

/////////////////////////////////////////////////////////////////////
// Helpers

func createTmpFile(t *testing.T, name string, content string) string {
	tempDir := t.TempDir()
	fname := filepath.Join(tempDir, name)
	if err := os.WriteFile(fname, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return fname
}
