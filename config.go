package config

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
)

const SourceCommandLine = "command line arguments"
const SourceEnvVars = "environment variables"
const CustomTagName = "param"

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

func (c *Config[ConfigType]) UsageFlags(out io.Writer) error {
	fs := flag.NewFlagSet(SourceCommandLine, flag.ContinueOnError)
	if err := buildFlagsForStruct(&c.cfg, fs); err != nil {
		return err
	}
	fs.SetOutput(out)
	fs.Usage()
	return nil
}

func (c *Config[ConfigType]) UsageEnvs(prefix string, out io.Writer) error {
	if out == nil {
		out = os.Stderr
	}
	fs := flag.NewFlagSet(SourceEnvVars, flag.ContinueOnError)
	if err := buildFlagsForStruct(&c.cfg, fs); err != nil {
		return err
	}
	fs.VisitAll(func(f *flag.Flag) {
		f.Name = buildEnvName(prefix, f.Name)
	})
	// custom writer cleans the dash heading the argument name
	fs.SetOutput(newPrefixCleanerWriter(out, []byte{' ', ' ', '-'}))
	fs.Usage()
	return nil
}

func (c *Config[ConfigType]) WithFile(
	file string,
	parser func(cfg *ConfigType, content []byte) error,
) *Config[ConfigType] {
	if c.err != nil {
		return c
	}
	content, err := os.ReadFile(file)
	if err != nil {
		c.err = err
		return c
	}
	c.err = parser(&c.cfg, content)
	return c
}

func (c *Config[ConfigType]) WithFlags(args []string) *Config[ConfigType] {
	if c.err != nil || len(args) == 0 {
		return c
	}
	fs := flag.NewFlagSet(SourceCommandLine, flag.ContinueOnError)
	if c.err = buildFlagsForStruct(&c.cfg, fs); c.err != nil {
		return c
	}
	fs.SetOutput(io.Discard) // avoid usage print
	if err := fs.Parse(args); err != nil {
		c.err = fmt.Errorf("failed to parse arguments: %v", err)
	}
	return c
}

func (c *Config[ConfigType]) WithEnvs(prefix string) *Config[ConfigType] {
	if c.err != nil {
		return c
	}
	fs := flag.NewFlagSet(SourceEnvVars, flag.ContinueOnError)
	if c.err = buildFlagsForStruct(&c.cfg, fs); c.err != nil {
		return c
	}
	fs.VisitAll(func(f *flag.Flag) {
		name := buildEnvName(prefix, f.Name)
		if value := os.Getenv(name); value != "" {
			if err := f.Value.Set(value); err != nil {
				c.err = fmt.Errorf("failed to parse %s (%s): %v", name, value, err)
			}
		}
	})
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
		return fmt.Errorf("enum: %s, value is not valid", prefix)
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		t := v.Type()
		if typesStack[t] {
			return fmt.Errorf("enum: %s, types recursion is not allowed", prefix)
		}
		typesStack[t] = true
		defer delete(typesStack, t)

		return enumerateValue(v.Elem(), prefix, typesStack, fs)
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("enum: %s, expected a struct, got %v", prefix, v.Kind())
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
		tag := parseCustomTag(structField.Tag.Get(CustomTagName))

		// Special handling for complex types
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			ptr := field.Addr().Interface().(*time.Duration)
			fs.DurationVar(ptr, name, *ptr, tag.Usage)
			continue
		}
		if field.Type().Implements(flagValueType) ||
			reflect.PointerTo(field.Type()).Implements(flagValueType) {
			ptr := field.Addr().Interface().(flag.Value)
			fs.Var(ptr, name, tag.Usage)
			continue
		}

		switch field.Kind() {
		case reflect.String:
			ptr := field.Addr().Interface().(*string)
			fs.StringVar(ptr, name, *ptr, tag.Usage)
		case reflect.Int:
			ptr := field.Addr().Interface().(*int)
			fs.IntVar(ptr, name, *ptr, tag.Usage)
		case reflect.Int64:
			ptr := field.Addr().Interface().(*int64)
			fs.Int64Var(ptr, name, *ptr, tag.Usage)
		case reflect.Uint:
			ptr := field.Addr().Interface().(*uint)
			fs.UintVar(ptr, name, *ptr, tag.Usage)
		case reflect.Uint64:
			ptr := field.Addr().Interface().(*uint64)
			fs.Uint64Var(ptr, name, *ptr, tag.Usage)
		case reflect.Bool:
			ptr := field.Addr().Interface().(*bool)
			fs.BoolVar(ptr, name, *ptr, tag.Usage)
		case reflect.Float64:
			ptr := field.Addr().Interface().(*float64)
			fs.Float64Var(ptr, name, *ptr, tag.Usage)
		case reflect.Struct, reflect.Ptr:
			if err := enumerateValue(field, name, typesStack, fs); err != nil {
				return err
			}
		default:
			return fmt.Errorf("enum: field %s, has unsupported type %v", name, field.Kind())
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

type customTag struct {
	Usage string
}

func parseCustomTag(tag string) customTag {
	return customTag{
		Usage: tag,
	}
}

type prefixCleanerWriter struct {
	w      io.Writer
	prefix []byte
}

func newPrefixCleanerWriter(w io.Writer, prefix []byte) *prefixCleanerWriter {
	return &prefixCleanerWriter{
		w:      w,
		prefix: prefix,
	}
}

func (w *prefixCleanerWriter) Write(p []byte) (n int, err error) {
	if bytes.HasPrefix(p, w.prefix) {
		for i := range w.prefix {
			p[i] = ' '
		}
	}
	return w.w.Write(p)
}
