[![Go Reference](https://pkg.go.dev/badge/github.com/serega-cpp/config.svg)](https://pkg.go.dev/github.com/serega-cpp/config)
[![Go Report Card](https://goreportcard.com/badge/github.com/serega-cpp/config)](https://goreportcard.com/report/github.com/serega-cpp/config)
[![Go Build](https://github.com/serega-cpp/config/actions/workflows/build.yaml/badge.svg)](https://github.com/serega-cpp/config/actions/workflows/build.yaml)
[![codecov](https://codecov.io/gh/serega-cpp/config/branch/master/graph/badge.svg)](https://codecov.io/gh/serega-cpp/config)

# Config

This package is designed to unify work with service configuration. It provides a convenient way to load configuration parameters into a data structure. When loading, they are converted to the appropriate Go data types. Loading from multiple sources is supported, and source chains can be built, where parameters are loaded sequentially from several sources, overwriting previous values. The package relies on the standard `package flag`, so there may be limitations and peculiarities associated with it.

### Parameter sources

The following sources are supported:
- files (via external packages: Yaml, Ini, Toml, etc.)
- command line parameters
- environment variables

Loading from a file is not directly the responsibility of this package; it merely provides the ability to implement a callback function in client code. Loading from command line parameters and environment variables, as well as config sources chaining, is implemented in the package.

### Supported parameter types

The list of Go types supported by file parsers depends on the specific parser implementation. The default set of types available for loading from command-line parameters and environment variables includes:
- string
- int, int64
- uint, uint64
- bool
- float64
- time.Duration

It is also supported to add custom types by initializing them from a string.

### Custom types

To use a custom type, you must implement the flag.Value interface for that type. This is sufficient for command line parsers and environment variables. See the example in the [custom_type_test.go](custom_type_test.go) or the theory https://pkg.go.dev/flag#example-Value.

To use a custom type with file parsers, follow the appropriate instructions for the chosen parser. See the example for a YAML file in the [config_yaml_test.go](config_yaml_test.go).

### Parameter names

Parameter names are automatically generated based on the structure field names. This is the package's key feature. For command line parameters, field names are first converted to lowercase and separated by hyphens. For environment variables, they are prefixed by the prefix, capitalized and separated by underscores. E.g.:

```
type Host struct {
	Addr string `usage:"Service IP listen on"`
	Port int    `usage:"Service port listen on"`
}

type Measurement struct {
	Duration time.Duration `usage:"Duration of experiment"`
}

type Config struct {
	Host        Host
	Measurement Measurement
}
```

produces the following command line parameter names:

```
--host-addr
--host-port
--measurement-duration
```

and the following environment names:

```
PREFIX_HOST_ADDR
PREFIX_HOST_PORT
PREFIX_MEASUREMENT_DURATION
```

In file parsers, parameter names are defined by the parser implementation itself. Additionally, they typically support tags for customizing names, which can be freely used here. This package does not support a tag for customizing names. It only supports the `usage` tag for describing values.

### Limitations

1. The pointers to data fields not supported, e.g.:
```
type BadConfig struct {
	Id *string
}
```

2. Recurrent types not supported, e.g.:
```
type BadConfig struct {
	Self *BadConfig
}
```

### Usage:

1. Create a configuration file, specifying initial values ​​if necessary
```
// if nil is provided as an initial struct, no defaults will be used
cfgObj := config.New(&Config{
	// ...
})
```

2. Load values ​​from file [optional]
```
cfgObj.WithFile(fileName,
	func(cfg *Config, content []byte) error {
		// here you should fill the cfg using the content value
	},
)
```

3. Load values from command line [optional]
```
// The argument is a slice with command line parameters,
// without program name (excluding the first item)
cfgObj.WithFlags(os.Args[1:])
```

4. Load values from environment variables [optional]
```
// The "prefix" is a prefix for all variable names,
// to avoid possible conflicts in a global enviroment space
objCfg.WithEnvs("prefix")
```

5. Prepare the configuration structure for use (and error if any)
```
cfg, err := objCfg.AsStruct()
```

6. You can also print out the usage information
```
// This prints the usage info for command line arguments, the argument
// is an output stream (if nil is provided, the stderr will be used)
cfgObj.UsageFlags(nil)

// This prints the usage info for environment variables, the 2-nd argument
// is an output stream (if nil is provided, the stderr will be used)
cfgObj.UsageEnvs("prefix", nil)
```

### Samples

**Sample #1:** Gets some defaults, then loads Yaml file, then parses command line and environment variables

```
package main

import (
	"github.com/serega-cpp/config"
	"gopkg.in/yaml.v3"
)

func main() {
	// This loads parameters in the order overwriting existing values:
	// defaults -> yaml file -> command line -> env vars
	// (calls may be freely reordered)

	cfg, err := config.New(&Config{
		Host: Host{
			Addr: "localhost",
		},
	}).WithFile("config.yaml",
		func(cfg *Config, content []byte) error {
			return yaml.Unmarshal(content, cfg)
		},
	).WithFlags(os.Args[1:]).WithEnvs("prefix").AsStruct()

	if err != nil {
		fmt.Println(err)
		return
	}

	// use cfg ...
}
```

**Sample #2:** Loads Ini-file content (using gopkg.in/ini.v1).

```
cfg, err := config.New[Config](nil).WithFile(fname,
	func(cfg *Config, content []byte) error {
		cnt, err := ini.Load(content)
		if err != nil {
			return err
		}
		return cnt.MapTo(cfg)
	},
).AsStruct()
```

**Sample #3:** Usage output.

```
config.New[Config](nil).UsageFlags(nil)
```

prints:

```
Usage of command line arguments:
  -host-addr string
    	Service IP listen on
  -host-port int
    	Service port listen on
  -measurement-duration duration
    	Duration of experiment
```

and

```
config.New[Config](nil).UsageEnvs("test", nil)
```

prints:

```
Usage of environment variables:
   TEST_HOST_ADDR string
    	Service IP listen on
   TEST_HOST_PORT int
    	Service port listen on
   TEST_MEASUREMENT_DURATION duration
    	Duration of experiment
```

**Sample #4:** Read config file name from command line.

Sometimes it's convenient to read the configuration file name from the command line. However, using such an option by the application will conflict with the main configuration loaded by the package. There is a solution: the "--" separator which correctly divides the groups of options (please note the usage of `flag.Args()` instead of `os.Args[1:]` for parsing next groups).

So, command line: `./app --config=sample.yaml -- --host-addr=127.0.0.1 --host-port=80 ...`

```
fname := flag.String("config", "config.yaml", "Configuration file")
flag.Parse()

cfg, err := config.New[Config](nil).WithFile(*fname,
	func(cfg *Config, content []byte) error {
		return yaml.Unmarshal(content, cfg)
	},
).WithFlags(flag.Args()).AsStruct()
```

Examples can also be found in [config_test.go](config_test.go).

### Installation

```
go get github.com/serega-cpp/config
```
