package main

import (
	"fmt"
	"os"

	"github.com/ormanli/mixingcheck/internal/check"
	"github.com/ormanli/mixingcheck/internal/config"
	"github.com/spf13/viper"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func initConfig() config.Packages {
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))
	v.AddConfigPath(".")
	v.SetConfigName(".mixingcheck")

	err := v.ReadInConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while reading config:", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, "Using config file:", v.ConfigFileUsed())

	var c config.Packages

	err = v.Unmarshal(&c)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while unmarshalling config:", err)
		os.Exit(1)
	}

	return c
}

func main() {
	c := initConfig()
	analyzer, err := check.NewAnalyzer(c)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while initializing analyzer:", err)
		os.Exit(1)
	}

	singlechecker.Main(analyzer)
}
