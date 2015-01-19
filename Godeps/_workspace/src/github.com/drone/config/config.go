package config

import (
	"flag"
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// -- ConfigSet

type ConfigSet struct {
	*flag.FlagSet
	prefix string
}

// SetPrefix sets the applicatin prefix to use when filtering environment variables.
// If prefix is empty, all environment variables are ignored.
func (c *ConfigSet) SetPrefix(prefix string) {
	c.prefix = prefix
}

// BoolVar defines a bool config with a given name and default value for a ConfigSet.
// The argument p points to a bool variable in which to store the value of the config.
func (c *ConfigSet) BoolVar(p *bool, name string, value bool) {
	c.FlagSet.BoolVar(p, name, value, "")
}

// Bool defines a bool config variable with a given name and default value for
// a ConfigSet.
func (c *ConfigSet) Bool(name string, value bool) *bool {
	return c.FlagSet.Bool(name, value, "")
}

// IntVar defines a int config with a given name and default value for a ConfigSet.
// The argument p points to a int variable in which to store the value of the config.
func (c *ConfigSet) IntVar(p *int, name string, value int) {
	c.FlagSet.IntVar(p, name, value, "")
}

// Int defines a int config variable with a given name and default value for a
// ConfigSet.
func (c *ConfigSet) Int(name string, value int) *int {
	return c.FlagSet.Int(name, value, "")
}

// Int64Var defines a int64 config with a given name and default value for a ConfigSet.
// The argument p points to a int64 variable in which to store the value of the config.
func (c *ConfigSet) Int64Var(p *int64, name string, value int64) {
	c.FlagSet.Int64Var(p, name, value, "")
}

// Int64 defines a int64 config variable with a given name and default value
// for a ConfigSet.
func (c *ConfigSet) Int64(name string, value int64) *int64 {
	return c.FlagSet.Int64(name, value, "")
}

// UintVar defines a uint config with a given name and default value for a ConfigSet.
// The argument p points to a uint variable in which to store the value of the config.
func (c *ConfigSet) UintVar(p *uint, name string, value uint) {
	c.FlagSet.UintVar(p, name, value, "")
}

// Uint defines a uint config variable with a given name and default value for
// a ConfigSet.
func (c *ConfigSet) Uint(name string, value uint) *uint {
	return c.FlagSet.Uint(name, value, "")
}

// Uint64Var defines a uint64 config with a given name and default value for a ConfigSet.
// The argument p points to a uint64 variable in which to store the value of the config.
func (c *ConfigSet) Uint64Var(p *uint64, name string, value uint64) {
	c.FlagSet.Uint64Var(p, name, value, "")
}

// Uint64 defines a uint64 config variable with a given name and default value
// for a ConfigSet.
func (c *ConfigSet) Uint64(name string, value uint64) *uint64 {
	return c.FlagSet.Uint64(name, value, "")
}

// StringVar defines a string config with a given name and default value for a ConfigSet.
// The argument p points to a string variable in which to store the value of the config.
func (c *ConfigSet) StringVar(p *string, name string, value string) {
	c.FlagSet.StringVar(p, name, value, "")
}

// String defines a string config variable with a given name and default value
// for a ConfigSet.
func (c *ConfigSet) String(name string, value string) *string {
	return c.FlagSet.String(name, value, "")
}

// Float64Var defines a float64 config with a given name and default value for a ConfigSet.
// The argument p points to a float64 variable in which to store the value of the config.
func (c *ConfigSet) Float64Var(p *float64, name string, value float64) {
	c.FlagSet.Float64Var(p, name, value, "")
}

// Float64 defines a float64 config variable with a given name and default
// value for a ConfigSet.
func (c *ConfigSet) Float64(name string, value float64) *float64 {
	return c.FlagSet.Float64(name, value, "")
}

// DurationVar defines a time.Duration config with a given name and default value for a ConfigSet.
// The argument p points to a time.Duration variable in which to store the value of the config.
func (c *ConfigSet) DurationVar(p *time.Duration, name string, value time.Duration) {
	c.FlagSet.DurationVar(p, name, value, "")
}

// Duration defines a time.Duration config variable with a given name and
// default value.
func (c *ConfigSet) Duration(name string, value time.Duration) *time.Duration {
	return c.FlagSet.Duration(name, value, "")
}

// Strings defines a string slice config variable with a given name.
func (c *ConfigSet) Strings(name string) *StringParams {
	p := new(StringParams)
	c.FlagSet.Var(p, name, "")
	return p
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value.
func (c *ConfigSet) Var(value flag.Value, name string) {
	c.FlagSet.Var(value, name, "")
}

// Parse takes a path to a TOML file and loads it. This must be called after
// all the config flags in the ConfigSet have been defined but before the flags
// are accessed by the program.
//
// If the path is empty, no TOML file is loaded. Only environment variables
// matching the application Prefix will populate the config flags.
func (c *ConfigSet) Parse(path string) error {
	data, err := ioutil.ReadFile(path)
	switch {
	case err != nil && len(path) != 0:
		return err
	case err != nil && len(path) == 0:
		data = []byte{}
	}
	return c.ParseBytes(data)
}

// ParseBytes takes a TOML file in byte array format and loads it. This must be
// called after all the config flags in the ConfigSet have been defined but before
// the flags are accessed by the program.
func (c *ConfigSet) ParseBytes(data []byte) error {
	tree, err := toml.Load(string(data))
	if err != nil {
		return err
	}

	err = c.loadTomlTree(tree, []string{})
	if err != nil {
		return err
	}

	if len(c.prefix) != 0 {
		envs := parseEnvVars(os.Environ())
		err = c.loadEnvVars(envs)
		if err != nil {
			return err
		}
	}

	return nil
}

// loadTomlTree recursively loads a TomlTree into this ConfigSet's config
// variables.
func (c *ConfigSet) loadTomlTree(tree *toml.TomlTree, path []string) error {
	for _, key := range tree.Keys() {
		fullPath := append(path, key)
		value := tree.Get(key)
		if subtree, isTree := value.(*toml.TomlTree); isTree {
			err := c.loadTomlTree(subtree, fullPath)
			if err != nil {
				return err
			}
		} else {
			fullPath := strings.Join(append(path, key), "-")
			fullPath = strings.Replace(fullPath, "_", "-", -1)
			switch v := value.(type) {
			case []interface{}:
				var items []string
				for _, item := range v {
					items = append(items, fmt.Sprintf("%v", item))
				}
				value = strings.Join(items, ",")
			}
			// TODO(bradrydzewski) handle []int, []int64, []float64, []float, etc
			// TODO(bradrydzewski) handle tables [[ ]] as map[interface{}]interface{}
			err := c.Set(fullPath, fmt.Sprintf("%v", value))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// loadEnvVars loads environment variables into thie ConfigSet's config
// variables if they match the use-defined PREFIX.
func (c *ConfigSet) loadEnvVars(environ map[string]string) error {
	for key, value := range environ {
		if !strings.HasPrefix(key, c.prefix) {
			continue
		}
		key = strings.ToLower(key)
		key = strings.Replace(key, "_", "-", -1)
		c.Set(key[len(c.prefix):], value)
	}
	return nil
}

// parseEnvVars parses a string of environment variables
// into a key-value map.
func parseEnvVars(environ []string) map[string]string {
	envs := map[string]string{}
	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		envs[parts[0]] = parts[1]
	}
	return envs
}

const (
	ContinueOnError flag.ErrorHandling = flag.ContinueOnError
	ExitOnError     flag.ErrorHandling = flag.ExitOnError
	PanicOnError    flag.ErrorHandling = flag.PanicOnError
)

// NewConfigSet returns a new ConfigSet with the given name and error handling
// policy. The three valid error handling policies are: flag.ContinueOnError,
// flag.ExitOnError, and flag.PanicOnError.
func NewConfigSet(name string, errorHandling flag.ErrorHandling) *ConfigSet {
	return &ConfigSet{
		FlagSet: flag.NewFlagSet(name, errorHandling),
	}
}

// -- globalConfig

var globalConfig = NewConfigSet(os.Args[0], flag.ExitOnError)

// SetPrefix sets the applicatin prefix to use when filtering environment variables.
// If prefix is empty, all environment variables are ignored.
func SetPrefix(prefix string) {
	globalConfig.prefix = prefix
}

// BoolVar defines a bool config with a given name and default value.
// The argument p points to a bool variable in which to store the value of the config.
func BoolVar(p *bool, name string, value bool) {
	globalConfig.BoolVar(p, name, value)
}

// Bool defines a bool config variable with a given name and default value.
func Bool(name string, value bool) *bool {
	return globalConfig.Bool(name, value)
}

// IntVar defines a int config with a given name and default value.
// The argument p points to a int variable in which to store the value of the config.
func IntVar(p *int, name string, value int) {
	globalConfig.IntVar(p, name, value)
}

// Int defines a int config variable with a given name and default value.
func Int(name string, value int) *int {
	return globalConfig.Int(name, value)
}

// Int64Var defines a int64 config with a given name and default value.
// The argument p points to a int64 variable in which to store the value of the config.
func Int64Var(p *int64, name string, value int64) {
	globalConfig.Int64Var(p, name, value)
}

// Int64 defines a int64 config variable with a given name and default value.
func Int64(name string, value int64) *int64 {
	return globalConfig.Int64(name, value)
}

// UintVar defines a uint config with a given name and default value.
// The argument p points to a uint variable in which to store the value of the config.
func UintVar(p *uint, name string, value uint) {
	globalConfig.UintVar(p, name, value)
}

// Uint defines a uint config variable with a given name and default value.
func Uint(name string, value uint) *uint {
	return globalConfig.Uint(name, value)
}

// Uint64Var defines a uint64 config with a given name and default value.
// The argument p points to a uint64 variable in which to store the value of the config.
func Uint64Var(p *uint64, name string, value uint64) {
	globalConfig.Uint64Var(p, name, value)
}

// Uint64 defines a uint64 config variable with a given name and default value.
func Uint64(name string, value uint64) *uint64 {
	return globalConfig.Uint64(name, value)
}

// StringVar defines a string config with a given name and default value.
// The argument p points to a string variable in which to store the value of the config.
func StringVar(p *string, name string, value string) {
	globalConfig.StringVar(p, name, value)
}

// String defines a string config variable with a given name and default value.
func String(name string, value string) *string {
	return globalConfig.String(name, value)
}

// Float64Var defines a float64 config with a given name and default value.
// The argument p points to a float64 variable in which to store the value of the config.
func Float64Var(p *float64, name string, value float64) {
	globalConfig.Float64Var(p, name, value)
}

// Float64 defines a float64 config variable with a given name and default
// value.
func Float64(name string, value float64) *float64 {
	return globalConfig.Float64(name, value)
}

// DurationVar defines a time.Duration config with a given name and default value.
// The argument p points to a time.Duration variable in which to store the value of the config.
func DurationVar(p *time.Duration, name string, value time.Duration) {
	globalConfig.DurationVar(p, name, value)
}

// Duration defines a time.Duration config variable with a given name and
// default value.
func Duration(name string, value time.Duration) *time.Duration {
	return globalConfig.Duration(name, value)
}

// Strings defines a string slice config variable with a given name.
func Strings(name string) *StringParams {
	return globalConfig.Strings(name)
}

// Var defines a flag with the specified name and usage string. The type and value of the
// flag are represented by the first argument, of type Value, which typically holds a
// user-defined implementation of Value.
func Var(value flag.Value, name string) {
	globalConfig.Var(value, name)
}

// Parse takes a path to a TOML file and loads it into the global ConfigSet.
// This must be called after all config flags have been defined but before the
// flags are accessed by the program.
func Parse(path string) error {
	return globalConfig.Parse(path)
}

// ParseBytes takes a TOML file in byte array format and loads it into the
// global ConfigSet. This must be called after all config flags have been defined
// but before the flags are accessed by the program.
func ParseBytes(path string) error {
	return globalConfig.Parse(path)
}

// -- Custom Types

type StringParams []string

func (s *StringParams) String() string {
	return fmt.Sprint(*s)
}

func (s *StringParams) Set(value string) error {
	for _, ss := range strings.Split(value, ",") {
		ss = strings.TrimSpace(ss)
		*s = append(*s, ss)
	}
	return nil
}
