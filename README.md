# mixingcheck

mixingcheck is a static analyzer to check forbidden structs and method calls.

## Installation

Execute:

```bash
$ go get github.com/ormanli/mixingcheck
```

Or using [Homebrew 🍺](https://brew.sh)

```bash
brew tap ormanli/mixingcheck https://github.com/ormanli/mixingcheck
brew install mixingcheck
```

Or download from [Releases](https://github.com/ormanli/mixingcheck/releases) page.

## How to use

Add `.mixingcheck.yaml` to your project.

```yaml
github.com/ormanli/mixingcheck: #1
  rules:
    - package: #2
        value: github.com/spf13/viper
      name:
        value: NewWithOptions
      type: call #3
    - package:
        value: github.com/ormanli/mixingcheck/internal/config
      name:
        value: Packages
      type: struct
github.com/ormanli/mixingcheck/internal: #4
  ignore_parent_rules: true #5
  rules:
    - package:
        value: regexp
      name:
        regex: true #6
        value: .*
      type: struct
```
1. Name of the package you want to execute rules.
2. The package that contains struct or function you want to check for rule match. 
3. Type is either struct or call.
4. It is possible to define additional rules for child packages.
5. If a is true, parent packages rules ignored.
6. If regex is true, value can be a regular expression.


Run mixingcheck, if there is any error it won't return zero.
```bash
$ mixingcheck .
Using config file: ~/ormanli/git/mixingcheck/.mixingcheck.yaml
~/ormanli/git/mixingcheck/main.go:13:19: hit struct rule github.com/ormanli/mixingcheck/internal/config.Packages
~/ormanli/git/mixingcheck/main.go:26:8: hit struct rule github.com/ormanli/mixingcheck/internal/config.Packages
~/ormanli/git/mixingcheck/main.go:14:7: hit call rule github.com/spf13/viper.NewWithOptions
```