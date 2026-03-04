package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
)

type Config[ConfigType any] struct {
	cfg ConfigType
	err error
}

func New[ConfigType any](initial *ConfigType) *Config[ConfigType] {
	if initial != nil {
		return &Config[ConfigType]{
			cfg: *initial,
		}
	}
	return &Config[ConfigType]{}
}

func (c *Config[ConfigType]) UsageFlags(out io.Writer) {
	fs := flag.NewFlagSet("command line arguments", flag.ContinueOnError)
	if buildFlagsForStruct(&c.cfg, fs) == nil {
		fs.SetOutput(out)
		fs.Usage()
	}
}

func (c *Config[ConfigType]) UsageEnvs(prefix string, out io.Writer) {
	if out == nil {
		out = os.Stderr
	}
	fs := flag.NewFlagSet(prefix, flag.ContinueOnError)
	if buildFlagsForStruct(&c.cfg, fs) == nil {
		fmt.Fprintln(out, "Usage of environment variables:")
		fs.VisitAll(func(f *flag.Flag) {
			name := buildEnvName(prefix, f.Name)
			fmt.Fprintf(out, "  %s\t\t%s\n", name, f.Usage)
		})
	}
}

func (c *Config[ConfigType]) WithFile(file string, parser func(content []byte) (ConfigType, error)) *Config[ConfigType] {
	if c.err != nil {
		return c
	}
	content, err := os.ReadFile(file)
	if err != nil {
		c.err = err
		return c
	}
	c.cfg, c.err = parser(content)
	return c
}

func (c *Config[ConfigType]) WithFlags(args []string, out io.Writer) *Config[ConfigType] {
	if c.err != nil || len(args) == 0 {
		return c
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	if c.err = buildFlagsForStruct(&c.cfg, fs); c.err == nil {
		fs.SetOutput(out)
		fs.Parse(args)
	}
	return c
}

func (c *Config[ConfigType]) WithEnvs(prefix string) *Config[ConfigType] {
	if c.err != nil {
		return c
	}
	fs := flag.NewFlagSet(prefix, flag.ContinueOnError)
	if c.err = buildFlagsForStruct(&c.cfg, fs); c.err == nil {
		fs.VisitAll(func(f *flag.Flag) {
			name := buildEnvName(prefix, f.Name)
			if value := os.Getenv(name); value != "" {
				f.Value.Set(value)
			}
		})
	}
	return c
}

func (c *Config[ConfigType]) AsStruct() (*ConfigType, error) {
	return &c.cfg, c.err
}

////////////////////////////////////////////////////////////////////////////
// Internals

var flagValueType reflect.Type = reflect.TypeOf((*flag.Value)(nil)).Elem()

func buildFlagsForStruct(s any, fs *flag.FlagSet) error {
	typesStack := make(map[reflect.Type]bool)
	return enumerateValue(reflect.ValueOf(s), "", typesStack, fs)
}

func enumerateValue(v reflect.Value, prefix string, typesStack map[reflect.Type]bool, fs *flag.FlagSet) error {
	if !v.IsValid() {
		return fmt.Errorf("enum: value is not valid")
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		t := v.Type()
		if typesStack[t] {
			return fmt.Errorf("enum: types recursion is not allowed")
		}
		typesStack[t] = true
		defer delete(typesStack, t)

		return enumerateValue(v.Elem(), prefix, typesStack, fs)
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct, got %v", v.Kind())
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := v.Type().Field(i)
		if !field.CanSet() || !field.CanAddr() {
			continue
		}
		if structField.PkgPath != "" {
			continue // skip unexported symbols
		}

		name := buildName(prefix, structField.Name)
		tagUsage := structField.Tag.Get("usage")

		// Special handling for complex types
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			ptr := field.Addr().Interface().(*time.Duration)
			fs.DurationVar(ptr, name, *ptr, tagUsage)
			continue
		}
		if field.Type().Implements(flagValueType) ||
			reflect.PointerTo(field.Type()).Implements(flagValueType) {
			ptr := field.Addr().Interface().(flag.Value)
			fs.Var(ptr, name, tagUsage)
			continue
		}

		switch field.Kind() {
		case reflect.String:
			ptr := field.Addr().Interface().(*string)
			fs.StringVar(ptr, name, *ptr, tagUsage)
		case reflect.Int:
			ptr := field.Addr().Interface().(*int)
			fs.IntVar(ptr, name, *ptr, tagUsage)
		case reflect.Int64:
			ptr := field.Addr().Interface().(*int64)
			fs.Int64Var(ptr, name, *ptr, tagUsage)
		case reflect.Uint:
			ptr := field.Addr().Interface().(*uint)
			fs.UintVar(ptr, name, *ptr, tagUsage)
		case reflect.Uint64:
			ptr := field.Addr().Interface().(*uint64)
			fs.Uint64Var(ptr, name, *ptr, tagUsage)
		case reflect.Bool:
			ptr := field.Addr().Interface().(*bool)
			fs.BoolVar(ptr, name, *ptr, tagUsage)
		case reflect.Float64:
			ptr := field.Addr().Interface().(*float64)
			fs.Float64Var(ptr, name, *ptr, tagUsage)
		case reflect.Struct, reflect.Ptr:
			if err := enumerateValue(field, name, typesStack, fs); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildName(prefix string, name string) string {
	if prefix == "" {
		return strings.ToLower(name)
	}
	return strings.ToLower(fmt.Sprintf("%s-%s", prefix, name))
}

func buildEnvName(prefix string, name string) string {
	s := prefix + "_" + strings.ReplaceAll(name, "-", "_")
	return strings.ToUpper(s)
}
