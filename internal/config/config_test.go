package config_test

import (
	"bytes"
	"testing"

	"github.com/ormanli/mixingcheck/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// Regression test for a bug where ignore_parent_rules in YAML never reached
// Package.IgnoreParentRules: Viper/mapstructure's default field matcher
// doesn't map a snake_case key to a CamelCase field without an explicit
// mapstructure tag. Unlike the check package's tests, which build
// config.Packages directly in Go, this goes through the same Viper decode
// path main.go uses, since that's the only way to catch this class of bug.
func Test_Unmarshal_IgnoreParentRules(t *testing.T) {
	yamlData := []byte(`
github.com/example/pkg:
  ignore_parent_rules: true
  rules:
    - package:
        value: sort
      name:
        value: Ints
      type: call
`)

	v := viper.NewWithOptions(viper.KeyDelimiter("::"))
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(bytes.NewReader(yamlData)))

	var c config.Packages
	require.NoError(t, v.Unmarshal(&c))

	pkg, ok := c["github.com/example/pkg"]
	require.True(t, ok)
	require.True(t, pkg.IgnoreParentRules)
	require.Len(t, pkg.Rules, 1)
}
