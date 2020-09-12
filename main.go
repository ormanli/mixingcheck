package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/ormanli/mixingcheck/internal/check"
	"github.com/ormanli/mixingcheck/internal/config"
	"github.com/spf13/viper"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func initConfig() (config.Packages, error) {
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))
	v.AddConfigPath(".")
	v.SetConfigName(".mixingcheck")

	var c config.Packages

	err := v.ReadInConfig()
	if err != nil {
		return c, fmt.Errorf("while reading config: %w", err)
	}

	fmt.Fprintln(os.Stdout, "Using config file:", v.ConfigFileUsed())

	err = v.Unmarshal(&c)
	if err != nil {
		return c, fmt.Errorf("while unmarshalling config: %w", err)
	}

	return c, nil
}

func main() {
	debug.SetGCPercent(-1)

	c, err := initConfig()
	analyzer := check.NewAnalyzer(c, err)

	singlechecker.Main(analyzer)
}
