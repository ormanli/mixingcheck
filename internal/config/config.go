package config

import (
	"fmt"
	"regexp"
)

// Packages defines rules package mapping.
type Packages map[string]Package

// Package defines rules for a package.
type Package struct {
	IgnoreParentRules bool
	Rules             []Rule
}

// Rule defines a rule.
type Rule struct {
	Type    RuleType
	Name    String
	Package String
}

// String is a wrapper around string is either plain string or regular expresion.
type String struct {
	Value  string
	Regex  bool
	regexp *regexp.Regexp
}

// Match checks if given string match.
// If it is regex, compiled regex used.
func (s *String) Match(m string) bool {
	if s.Regex {
		return s.regexp.MatchString(m)
	}

	return s.Value == m
}

// Compile compiles String if it is a regular expression.
func (s *String) Compile(pattern string) error {
	if s.Regex {
		r, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}

		s.regexp = r
	}

	return nil
}

func (s *String) String() string {
	if s.Regex {
		return s.regexp.String()
	}

	return s.Value
}

func (r *Rule) String() string {
	return fmt.Sprintf("hit %s rule %s.%s", r.Type, r.Package.String(), r.Name.String())
}

// RuleType defines rule type.
type RuleType string

const (
	// StructRule defines rule type of struct definitions.
	StructRule = "struct"
	// CallRule defines rule type of function calls.
	CallRule = "call"
)
