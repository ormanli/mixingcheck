package main

import (
	"fmt"
	"os"

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
	c, err := initConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error", err)
		os.Exit(1)
	}

	analyzer, err := check.NewAnalyzer(c)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error", err)
		os.Exit(1)
	}

	singlechecker.Main(analyzer)
}
