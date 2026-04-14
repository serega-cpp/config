package config_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/serega-cpp/config"
	"github.com/stretchr/testify/require"
)

/////////////////////////////////////////////////////////////////
// The package supports 3 types of the config struct declaration:
// 1. With sub-structs as a values (classic)
// 2. With pointers to sub-structs (pointers)
// 3. With embedded, anonymous sub-structs (embedded)
//
// The test includes all supported datatypes.

type Simple struct {
	Str     string  `usage:"::str field"`
	Int     int     `usage:"::int field"`
	Uint    uint    `usage:"::uint field"`
	Bool    bool    `usage:"::bool field"`
	Float64 float64 `usage:"::float64 field"`
	Int64   int64   `usage:"::int64 field"`
	Uint64  uint64  `usage:"::uint64 field"`
	ignored string  `usage:"unexported field"`
}
type External struct {
	Duration time.Duration `usage:"::duration field"`
}

// Classic config with values
type MyConfig struct {
	Id       string `usage:"Identificator"`
	Simple   Simple
	External External
}

// Config with pointers
type MyConfigPtr struct {
	Id       string `usage:"Identificator"`
	Simple   *Simple
	External *External
}

// Config with anonymous embedded sub-structs
type MyConfigEmbed struct {
	Id     string `usage:"Identificator"`
	Simple struct {
		Str     string  `usage:"::str field"`
		Int     int     `usage:"::int field"`
		Uint    uint    `usage:"::uint field"`
		Bool    bool    `usage:"::bool field"`
		Float64 float64 `usage:"::float64 field"`
		Int64   int64   `usage:"::int64 field"`
		Uint64  uint64  `usage:"::uint64 field"`
		ignored string  `usage:"unexported field"`
	}
	External struct {
		Duration time.Duration `usage:"::duration field"`
	}
}

// Various unallowed configs
type BadConfig struct {
	Id *string
}
type RecurrentConfig struct {
	Self *RecurrentConfig
}
type InvalidConfig struct {
	Iface interface{}
}

///////////////////////////////////////////////////////////
// Sample config content

var argsSample = []string{
	"--id=test",
	"--simple-str=BMW",
	"--simple-int=-8",
	"--simple-uint=1990",
	"--simple-bool=true",
	"--simple-float64=3.14",
	"--simple-int64=-6123456789",
	"--simple-uint64=8123456789",
	"--external-duration=12s",
}

var envsSample = []string{
	"TEST_ID=test",
	"TEST_SIMPLE_STR=BMW",
	"TEST_SIMPLE_INT=-8",
	"TEST_SIMPLE_UINT=1990",
	"TEST_SIMPLE_BOOL=true",
	"TEST_SIMPLE_FLOAT64=3.14",
	"TEST_SIMPLE_INT64=-6123456789",
	"TEST_SIMPLE_UINT64=8123456789",
	"TEST_EXTERNAL_DURATION=12s",
}

///////////////////////////////////////////////////////////
// Config parts expected

var idExpected = "test"
var simpleExpected = Simple{
	Str:     "BMW",
	Int:     -8,
	Uint:    1990,
	Bool:    true,
	Float64: 3.14,
	Int64:   -6123456789,
	Uint64:  8123456789,
}
var externalExpected = External{
	Duration: 12 * time.Second,
}

const usageFlagsExpected = "Usage of command line arguments:\n" +
	"  -external-duration duration\n" +
	"    \t::duration field\n" +
	"  -id string\n" +
	"    \tIdentificator\n" +
	"  -simple-bool\n" +
	"    \t::bool field\n" +
	"  -simple-float64 float\n" +
	"    \t::float64 field\n" +
	"  -simple-int int\n" +
	"    \t::int field\n" +
	"  -simple-int64 int\n" +
	"    \t::int64 field\n" +
	"  -simple-str string\n" +
	"    \t::str field\n" +
	"  -simple-uint uint\n" +
	"    \t::uint field\n" +
	"  -simple-uint64 uint\n" +
	"    \t::uint64 field\n"

const usageEnvsExpected = "Usage of environment variables:\n" +
	"   TEST_EXTERNAL_DURATION duration\n" +
	"    \t::duration field\n" +
	"   TEST_ID string\n" +
	"    \tIdentificator\n" +
	"   TEST_SIMPLE_BOOL\n" +
	"    \t::bool field\n" +
	"   TEST_SIMPLE_FLOAT64 float\n" +
	"    \t::float64 field\n" +
	"   TEST_SIMPLE_INT int\n" +
	"    \t::int field\n" +
	"   TEST_SIMPLE_INT64 int\n" +
	"    \t::int64 field\n" +
	"   TEST_SIMPLE_STR string\n" +
	"    \t::str field\n" +
	"   TEST_SIMPLE_UINT uint\n" +
	"    \t::uint field\n" +
	"   TEST_SIMPLE_UINT64 uint\n" +
	"    \t::uint64 field\n"

/////////////////////////////////////////////////////////
// Tests implementations

func TestConfigUsageFlags(t *testing.T) {
	t.Run("Usage flags (classic)", func(t *testing.T) {
		var buf bytes.Buffer
		config.New[MyConfig](nil).UsageFlags(&buf)
		require.Equal(t, usageFlagsExpected, buf.String())
	})
	t.Run("Usage flags (pointers)", func(t *testing.T) {
		var buf bytes.Buffer
		config.New[MyConfigPtr](nil).UsageFlags(&buf)
		require.Equal(t, usageFlagsExpected, buf.String())
	})
	t.Run("Usage flags (embedded)", func(t *testing.T) {
		var buf bytes.Buffer
		config.New[MyConfigEmbed](nil).UsageFlags(&buf)
		require.Equal(t, usageFlagsExpected, buf.String())
	})
}

func TestConfigFlags(t *testing.T) {
	t.Run("Flags (classic)", func(t *testing.T) {
		cfg, err := config.New[MyConfig](nil).WithFlags(argsSample).AsStruct()
		require.NoError(t, err)
		require.Equal(t, idExpected, cfg.Id)
		require.Equal(t, simpleExpected, cfg.Simple)
		require.Equal(t, externalExpected, cfg.External)
	})
	t.Run("Flags (pointers)", func(t *testing.T) {
		cfg, err := config.New[MyConfigPtr](nil).WithFlags(argsSample).AsStruct()
		require.NoError(t, err)
		require.Equal(t, idExpected, cfg.Id)
		require.Equal(t, &simpleExpected, cfg.Simple)
		require.Equal(t, &externalExpected, cfg.External)
	})
	t.Run("Flags (embedded)", func(t *testing.T) {
		cfg, err := config.New[MyConfigEmbed](nil).WithFlags(argsSample).AsStruct()
		require.NoError(t, err)
		require.Equal(t, idExpected, cfg.Id)
		require.Equal(t, simpleExpected, Simple(cfg.Simple))
		require.Equal(t, externalExpected, External(cfg.External))
	})
}

func TestConfigUsageEnvs(t *testing.T) {
	t.Run("Usage envs (classic)", func(t *testing.T) {
		err := config.New[MyConfig](nil).UsageEnvs("test", nil)
		require.NoError(t, err)

		var buf bytes.Buffer
		config.New[MyConfig](nil).UsageEnvs("test", &buf)
		require.Equal(t, usageEnvsExpected, buf.String())
	})
	t.Run("Usage envs (pointers)", func(t *testing.T) {
		err := config.New[MyConfigPtr](nil).UsageEnvs("test", nil)
		require.NoError(t, err)

		var buf bytes.Buffer
		config.New[MyConfigPtr](nil).UsageEnvs("test", &buf)
		require.Equal(t, usageEnvsExpected, buf.String())
	})
	t.Run("Usage envs (embedded)", func(t *testing.T) {
		err := config.New[MyConfigEmbed](nil).UsageEnvs("test", nil)
		require.NoError(t, err)

		var buf bytes.Buffer
		config.New[MyConfigEmbed](nil).UsageEnvs("test", &buf)
		require.Equal(t, usageEnvsExpected, buf.String())
	})
}

func TestConfigEnvs(t *testing.T) {
	createEnvironment(t, envsSample)
	t.Run("Envs (classic)", func(t *testing.T) {
		cfg, err := config.New[MyConfig](nil).WithEnvs("test").AsStruct()
		require.NoError(t, err)
		require.Equal(t, idExpected, cfg.Id)
		require.Equal(t, simpleExpected, cfg.Simple)
		require.Equal(t, externalExpected, cfg.External)
	})
	t.Run("Envs (pointers)", func(t *testing.T) {
		cfg, err := config.New[MyConfigPtr](nil).WithEnvs("test").AsStruct()
		require.NoError(t, err)
		require.Equal(t, idExpected, cfg.Id)
		require.Equal(t, &simpleExpected, cfg.Simple)
		require.Equal(t, &externalExpected, cfg.External)
	})
	t.Run("Envs (embedded)", func(t *testing.T) {
		cfg, err := config.New[MyConfigEmbed](nil).WithEnvs("test").AsStruct()
		require.NoError(t, err)
		require.Equal(t, idExpected, cfg.Id)
		require.Equal(t, simpleExpected, Simple(cfg.Simple))
		require.Equal(t, externalExpected, External(cfg.External))
	})
}

func TestConfigChain(t *testing.T) {
	var args = []string{
		"--simple-int=2",
		"--simple-float64=2.0",
	}
	var envs = []string{
		"CHAIN_SIMPLE_FLOAT64=3.0",
	}
	createEnvironment(t, envs)
	t.Run("Test chain", func(t *testing.T) {
		cfg, err := config.New(&MyConfig{
			Simple: Simple{
				Str:     "1",
				Int:     1,
				Float64: 1.0,
			},
		}).WithFlags(args).WithEnvs("chain").AsStruct()
		require.NoError(t, err)
		require.Equal(t, "1", cfg.Simple.Str)
		require.Equal(t, 2, cfg.Simple.Int)
		require.Equal(t, 3.0, cfg.Simple.Float64)
	})
}

func TestConfigErrors(t *testing.T) {
	t.Run("Test BadStruct usage", func(t *testing.T) {
		err1 := config.New[BadConfig](nil).UsageFlags(nil)
		require.Error(t, err1)
		err2 := config.New[BadConfig](nil).UsageEnvs("test", nil)
		require.Error(t, err2)
	})
	t.Run("Test BadStructs", func(t *testing.T) {
		_, err1 := config.New[BadConfig](nil).WithFlags([]string{"test"}).AsStruct()
		require.Error(t, err1)
		_, err2 := config.New[RecurrentConfig](nil).WithFlags([]string{"test"}).AsStruct()
		require.Error(t, err2)
		_, err3 := config.New[InvalidConfig](nil).WithFlags([]string{"test"}).AsStruct()
		require.Error(t, err3)
	})
	t.Run("Test withFile", func(t *testing.T) {
		_, err := config.New[MyConfig](nil).WithFile("nonexist",
			func(cfg *MyConfig, content []byte) error {
				return nil
			},
		).WithFile("nonexist",
			func(cfg *MyConfig, content []byte) error {
				require.Fail(t, "reads file after failure")
				return nil
			},
		).WithFlags(argsSample).WithEnvs("test").AsStruct()
		require.Error(t, err)
	})
}

/////////////////////////////////////////////////////////////////////
// Helpers

func createEnvironment(t *testing.T, envs []string) {
	for _, env := range envs {
		pos := strings.Index(env, "=")
		if pos == -1 {
			t.Fatalf("failed to parse env var: %s", env)
		}
		os.Setenv(env[:pos], env[pos+1:])
	}
}
