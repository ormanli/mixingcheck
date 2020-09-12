package check

import (
	"testing"

	"github.com/ormanli/mixingcheck/internal/config"
	"github.com/stretchr/testify/suite"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

type CheckerTestSuite struct {
	suite.Suite
	dir      string
	cleanup  func()
	files    map[string]map[string]string
	config   map[string]config.Packages
	analyzer *analysis.Analyzer
}

func TestCheckerTestSuite(t *testing.T) {
	suite.Run(t, new(CheckerTestSuite))
}

func (s *CheckerTestSuite) BeforeTest(_, testName string) {
	filemap := s.files[testName]
	s.Require().NotEmpty(filemap)

	dir, cleanup, err := analysistest.WriteFiles(filemap)
	s.Require().NoError(err)

	s.dir = dir
	s.cleanup = cleanup

	c := s.config[testName]
	s.Require().NotEmpty(c)

	s.analyzer = NewAnalyzer(c, nil)
}

func (s *CheckerTestSuite) SetupSuite() {
	s.files = make(map[string]map[string]string)
	s.config = make(map[string]config.Packages)

	s.files["Test_Basic"] = map[string]string{"a/main.go": `
package main

import (
	"fmt"
	"a/c"
)

func main() {
	ints := c.A()
	cda := c.CDA{ // want "hit struct rule a/c.CDA"
		E:   "test",
		Val: ints,
	}

	fmt.Printf("%s %v", cda.E, cda.Val)
}
`,
		"a/c/c.go": `
package c

import (
	"sort"
)

type CDA struct {
	E   string
	Val []int
}

func A() []int {
	ints := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}

	sort.Ints(ints) // want "hit call rule sort.Ints"

	return ints
}
`,
		"a/c/c_test.go": `
package c_test

import (
	"sort"
	"a/c"
	"testing"
)

func Test_A(t *testing.T) {
	data:=c.A()
	if !sort.IntsAreSorted(data) {
		t.Errorf("sort didn't sort")
	}
}
`}
	s.config["Test_Basic"] = config.Packages{"a": config.Package{
		Rules: []config.Rule{
			{
				Type: config.CallRule,
				Name: config.String{
					Value: "Ints",
				},
				Package: config.String{
					Value: "sort",
				},
			},
			{
				Type: config.StructRule,
				Name: config.String{
					Value: "CDA",
				},
				Package: config.String{
					Value: "a/c",
				},
			},
		},
	}}

	s.files["Test_Regex"] = map[string]string{"a/main.go": `
package main

import (
	"fmt"
	"a/c"
)

func main() {
	ints := c.A()
	cda := c.CDA{
		E:   "test",
		Val: ints,
	}

	fmt.Printf("%s %v", cda.E, cda.Val)
}
`,
		"a/c/c.go": `
package c

import (
	"sort"
	another_sort "x/y/sort"
)

type CDA struct {
	E   string
	Val []int
}

func A() []int {
	ints := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}

	sort.Ints(ints) // want "hit call rule .*.Ints"

	another_sort.Ints(ints) // want "hit call rule .*.Ints"

	return ints
}
`,
		"a/c/c_test.go": `
package c_test

import (
	"sort"
	"a/c"
	"testing"
)

func Test_A(t *testing.T) {
	data:=c.A()
	if !sort.IntsAreSorted(data) {
		t.Errorf("sort didn't sort")
	}
}
`,
		"x/y/sort/sort.go": `
package sort

func Ints(ints []int) {
}
`}
	s.config["Test_Regex"] = config.Packages{"a": config.Package{
		Rules: []config.Rule{
			{
				Type: config.CallRule,
				Name: config.String{
					Value: "Ints",
				},
				Package: config.String{
					Regex: true,
					Value: ".*",
				},
			},
		},
	}}

	s.files["Test_IgnoreParent"] = map[string]string{"a/main.go": `
package main

import (
	"fmt"
	"a/c"
)

func main() {
	ints := c.A()
	cda := c.CDA{ // want "hit struct rule a/c.CDA"
		E:   "test",
		Val: ints,
	}

	fmt.Printf("%s %v", cda.E, cda.Val)
}
`,
		"a/c/c.go": `
package c

import (
	"sort"
)

type CDA struct {
	E   string
	Val []int
}

func A() []int {
	ints := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}

	sort.Ints(ints)

	return ints
}
`,
		"a/c/c_test.go": `
package c_test

import (
	"sort"
	"a/c"
	"testing"
)

func Test_A(t *testing.T) {
	data:=c.A()
	if !sort.IntsAreSorted(data) {
		t.Errorf("sort didn't sort")
	}
}
`}
	s.config["Test_IgnoreParent"] = config.Packages{
		"a": config.Package{
			Rules: []config.Rule{
				{
					Type: config.CallRule,
					Name: config.String{
						Value: "Ints",
					},
					Package: config.String{
						Value: "sort",
					},
				},
				{
					Type: config.StructRule,
					Name: config.String{
						Value: "CDA",
					},
					Package: config.String{
						Value: "a/c",
					},
				},
			},
		},
		"a/c": config.Package{
			IgnoreParentRules: true,
		},
	}

	s.files["Test_Test"] = map[string]string{"a/main.go": `
package main

import (
	"fmt"
	"a/c"
)

func main() {
	ints := c.A()
	cda := c.CDA{
		E:   "test",
		Val: ints,
	}

	fmt.Printf("%s %v", cda.E, cda.Val)
}
`,
		"a/c/c.go": `
package c

import (
	"sort"
)

type CDA struct {
	E   string
	Val []int
}

func A() []int {
	ints := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}

	sort.Ints(ints)

	return ints
}
`,
		"a/c/c_test.go": `
package c_test

import (
	"sort"
	"a/c"
	"testing"
)

func Test_A(t *testing.T) {
	data:=c.A() // want "hit call rule a/c.A"
	if !sort.IntsAreSorted(data) {
		t.Errorf("sort didn't sort")
	}
}
`}
	s.config["Test_Test"] = config.Packages{
		"a/c_test": config.Package{
			Rules: []config.Rule{
				{
					Type: config.CallRule,
					Name: config.String{
						Value: "A",
					},
					Package: config.String{
						Value: "a/c",
					},
				},
			},
		},
	}
}

func (s *CheckerTestSuite) TearDownTest() {
	s.cleanup()
}

func (s *CheckerTestSuite) Test_Basic() {
	got := analysistest.Run(s.T(), s.dir, s.analyzer, "a", "a/c")
	s.Require().NotEmpty(got)

	var result []string
	for i := range got {
		for j := range got[i].Diagnostics {
			result = append(result, got[i].Diagnostics[j].Message)
		}
	}

	s.Require().ElementsMatch(result, []string{"hit struct rule a/c.CDA", "hit call rule sort.Ints"})
}

func (s *CheckerTestSuite) Test_Regex() {
	got := analysistest.Run(s.T(), s.dir, s.analyzer, "a", "a/c")
	s.Require().NotEmpty(got)

	var result []string
	for i := range got {
		for j := range got[i].Diagnostics {
			result = append(result, got[i].Diagnostics[j].Message)
		}
	}

	s.Require().ElementsMatch(result, []string{"hit call rule .*.Ints", "hit call rule .*.Ints"})
}

func (s *CheckerTestSuite) Test_IgnoreParent() {
	got := analysistest.Run(s.T(), s.dir, s.analyzer, "a", "a/c")
	s.Require().NotEmpty(got)

	var result []string
	for i := range got {
		for j := range got[i].Diagnostics {
			result = append(result, got[i].Diagnostics[j].Message)
		}
	}

	s.Require().ElementsMatch(result, []string{"hit struct rule a/c.CDA"})
}

func (s *CheckerTestSuite) Test_Test() {
	got := analysistest.Run(s.T(), s.dir, s.analyzer, "a", "a/c", "a/c.test")
	s.Require().NotEmpty(got)

	var result []string
	for i := range got {
		for j := range got[i].Diagnostics {
			result = append(result, got[i].Diagnostics[j].Message)
		}
	}

	s.Require().ElementsMatch(result, []string{"hit call rule a/c.A"})
}
