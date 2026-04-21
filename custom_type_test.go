package config_test

import (
	"bytes"
	"net/url"
	"testing"
	"time"

	"github.com/serega-cpp/config"
	"github.com/stretchr/testify/require"
)

// The Time type is used here to show how to implement customly parsed
// types. In short, to support the flags and environments variables as
// a source, you need to implement `flag.Value` interface.

type Time struct {
	time.Time
}

func (t *Time) String() string {
	return t.Time.String()
}

func (t *Time) Set(s string) error {
	tt, err := config.ParseTime(s)
	if err != nil {
		return err
	}
	t.Time = tt
	return nil
}

// ... and URL as a very popular type in configs)

type Url struct {
	*url.URL
}

func (u *Url) String() string {
	if u.URL == nil {
		return ""
	}
	return u.URL.String()
}

func (u *Url) Set(s string) error {
	uu, err := config.ParseUrl(s)
	if err != nil {
		return err
	}
	u.URL = uu
	return nil
}

////////////////////////////////////////////////////////
// Config struct declaration

type Custom struct {
	StdTime   Time `param:"::time20 field"`
	LongTime  Time `param:"::time19 field"`
	ShortTime Time `param:"::time10 field"`
	Link      Url  `param:"::Link field"`
}

type CustomConfig struct {
	Id     string `param:"Identificator"`
	Custom Custom
}

////////////////////////////////////////////////////////
// Sample config content

var argsCustom = []string{
	"--id=custom_type",
	"--custom-stdtime=2023-02-16T12:00:00Z",
	"--custom-longtime=2023-02-16T18:00:00",
	"--custom-shorttime=2025-12-21",
	"--custom-link=https://google.com/search?q=golang",
}

var envsCustom = []string{
	"TEST_ID=custom_type",
	"TEST_CUSTOM_STDTIME=2023-02-16T12:00:00Z",
	"TEST_CUSTOM_LONGTIME=2023-02-16T18:00:00",
	"TEST_CUSTOM_SHORTTIME=2025-12-21",
	"TEST_CUSTOM_LINK=https://google.com/search?q=golang",
}

var badArgsCustom = []string{
	"--custom-stdtime=2023-02-16T12:00",
}

var badEnvsCustom = []string{
	"TEST_CUSTOM_STDTIME=2023-02-16T12:00",
}

///////////////////////////////////////////////////////////
// Config expected

var idCustomExpected = "custom_type"
var customExpected = Custom{
	StdTime:   Time{time.Date(2023, 2, 16, 12, 0, 0, 0, time.UTC)},
	LongTime:  Time{time.Date(2023, 2, 16, 18, 0, 0, 0, time.UTC)},
	ShortTime: Time{time.Date(2025, 12, 21, 0, 0, 0, 0, time.UTC)},
	Link:      initUrl("https://google.com/search?q=golang"),
}

const usageFlagsCustomExpected = "Usage of command line arguments:\n" +
	"  -custom-link value\n" +
	"    \t::Link field\n" +
	"  -custom-longtime value\n" +
	"    \t::time19 field\n" +
	"  -custom-shorttime value\n" +
	"    \t::time10 field\n" +
	"  -custom-stdtime value\n" +
	"    \t::time20 field\n" +
	"  -id string\n" +
	"    \tIdentificator\n"

const usageEnvsCustomExpected = "Usage of environment variables:\n" +
	"   TEST_CUSTOM_LINK value\n" +
	"    \t::Link field\n" +
	"   TEST_CUSTOM_LONGTIME value\n" +
	"    \t::time19 field\n" +
	"   TEST_CUSTOM_SHORTTIME value\n" +
	"    \t::time10 field\n" +
	"   TEST_CUSTOM_STDTIME value\n" +
	"    \t::time20 field\n" +
	"   TEST_ID string\n" +
	"    \tIdentificator\n"

/////////////////////////////////////////////////////////
// Tests implementations

func TestConfigUsageCustomFlags(t *testing.T) {
	t.Run("Usage custom flags", func(t *testing.T) {
		var buf bytes.Buffer
		config.New[CustomConfig](nil).UsageFlags(&buf)
		require.Equal(t, usageFlagsCustomExpected, buf.String())
	})
}

func TestConfigCustomFlags(t *testing.T) {
	t.Run("Flags custom", func(t *testing.T) {
		cfg, err := config.New[CustomConfig](nil).WithFlags(argsCustom).AsStruct()
		require.NoError(t, err)
		require.Equal(t, idCustomExpected, cfg.Id)
		require.Equal(t, customExpected, cfg.Custom)
	})
}

func TestConfigUsageCustomEnvs(t *testing.T) {
	t.Run("Usage custom envs", func(t *testing.T) {
		var buf bytes.Buffer
		config.New[CustomConfig](nil).UsageEnvs("test", &buf)
		require.Equal(t, usageEnvsCustomExpected, buf.String())
	})
}

func TestConfigCustomEnvs(t *testing.T) {
	createEnvironment(t, envsCustom)
	t.Run("Envs custom", func(t *testing.T) {
		cfg, err := config.New[CustomConfig](nil).WithEnvs("test").AsStruct()
		require.NoError(t, err)
		require.Equal(t, idCustomExpected, cfg.Id)
		require.Equal(t, customExpected, cfg.Custom)
	})
}

func TestConfigBadCustomFlags(t *testing.T) {
	t.Run("Bad custom flags", func(t *testing.T) {
		_, err := config.New[CustomConfig](nil).WithFlags(badArgsCustom).AsStruct()
		require.Error(t, err)
	})
}

func TestConfigBadCustomEnvs(t *testing.T) {
	createEnvironment(t, badEnvsCustom)
	t.Run("Bad custom envs", func(t *testing.T) {
		_, err := config.New[CustomConfig](nil).WithEnvs("test").AsStruct()
		require.Error(t, err)
	})
}

/////////////////////////////////////////////////////////////////////
// Helpers

func initUrl(s string) Url {
	u, err := url.Parse(s)
	if err != nil {
		return Url{}
	}
	return Url{u}
}
