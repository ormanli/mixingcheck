package check

import (
	"bytes"
	"fmt"
	"go/ast"
	"strings"

	iradix "github.com/hashicorp/go-immutable-radix"
	"github.com/ormanli/mixingcheck/internal/config"
	"golang.org/x/tools/go/analysis"
)

func initializeTree(c config.Packages) (*iradix.Tree, error) {
	r := iradix.New()

	for name, pkg := range c {
		for i := range pkg.Rules {
			if !(pkg.Rules[i].Type == config.StructRule || pkg.Rules[i].Type == config.CallRule) {
				return nil, fmt.Errorf("rule %d in %s has invalid type", i, pkg.Rules[i].Type)
			}

			if pkg.Rules[i].Package.Value == "" {
				return nil, fmt.Errorf("rule %d in %s has empty package", i, name)
			}

			err := pkg.Rules[i].Package.Compile(pkg.Rules[i].Package.Value)
			if err != nil {
				return nil, fmt.Errorf("rule %d in %s has invalid regex, %w", i, name, err)
			}

			if pkg.Rules[i].Name.Value == "" {
				return nil, fmt.Errorf("rule %d in %s has empty package", i, name)
			}

			err = pkg.Rules[i].Name.Compile(pkg.Rules[i].Package.Value)
			if err != nil {
				return nil, fmt.Errorf("rule %d in %s has invalid regex, %w", i, name, err)
			}
		}

		r, _, _ = r.Insert([]byte(name), pkg)
	}

	return r, nil
}

// NewAnalyzer initializes analysis.Analyzer with given configuration.
func NewAnalyzer(c config.Packages) (*analysis.Analyzer, error) {
	tree, err := initializeTree(c)
	if err != nil {
		return nil, err
	}

	runner := runner{tree: tree}

	analyzer := &analysis.Analyzer{
		Name: "mixingcheck",
		Doc:  "mixingcheck",
		Run:  runner.run,
	}

	return analyzer, nil
}

type runner struct {
	tree *iradix.Tree
}

func (r runner) extractImports(file *ast.File) map[string]string {
	m := make(map[string]string)

	for _, importSpec := range file.Imports {
		path := strings.Trim(importSpec.Path.Value, `"`)

		if importSpec.Name != nil {
			m[importSpec.Name.Name] = path
		} else {
			name := path[strings.LastIndex(path, "/")+1:]
			name = strings.Split(name, ".")[0]
			m[name] = path
		}
	}

	return m
}

func (r *runner) run(pass *analysis.Pass) (interface{}, error) {
	rules := r.gatherRules(pass.Pkg.Path())

	for _, file := range pass.Files {
		importMap := r.extractImports(file)

		ast.Inspect(file, func(n ast.Node) bool {
			if cl, ok := n.(*ast.SelectorExpr); ok {
				diagnostics := checkSelector(cl, rules, importMap)
				for _, diagnostic := range diagnostics {
					pass.Report(diagnostic)
				}

				return true
			}

			return true
		})

		ast.Inspect(file, func(n ast.Node) bool {
			if ce, ok := n.(*ast.CallExpr); ok {
				diagnostics := checkCall(ce, rules, importMap)
				for _, diagnostic := range diagnostics {
					pass.Report(diagnostic)
				}

				return true
			}

			return true
		})
	}

	return nil, nil
}

func (r *runner) gatherRules(packageName string) []config.Rule {
	packageNameAsBytes := []byte(packageName)
	var otherRules []config.Rule
	var baseRules []config.Rule

	value, exist := r.tree.Root().Get(packageNameAsBytes)
	if exist {
		pkg := value.(config.Package)

		if pkg.IgnoreParentRules {
			return pkg.Rules
		}

		baseRules = pkg.Rules
	}

	r.tree.Root().WalkPath(packageNameAsBytes, func(k []byte, v interface{}) bool {
		if bytes.Equal(k, packageNameAsBytes) {
			return true
		}

		pkg := v.(config.Package)

		if pkg.IgnoreParentRules {
			otherRules = nil
		}

		otherRules = append(otherRules, pkg.Rules...)

		return false
	})

	for i, j := 0, len(otherRules)-1; i < j; i, j = i+1, j-1 {
		otherRules[i], otherRules[j] = otherRules[j], otherRules[i]
	}

	return append(baseRules, otherRules...)
}

func checkSelector(se *ast.SelectorExpr, rules []config.Rule, importMap map[string]string) []analysis.Diagnostic {
	var result []analysis.Diagnostic

	pkg, ok := se.X.(*ast.Ident)
	if !ok {
		return nil
	}

	for _, rule := range rules {
		if rule.Type == config.StructRule {
			match1 := rule.Name.Match(se.Sel.Name)

			match2 := rule.Package.Match(importMap[pkg.Name])

			if match1 && match2 {
				result = append(result, analysis.Diagnostic{
					Pos:     se.Pos(),
					Message: rule.String(),
				})
			}
		}
	}

	return result
}

func checkCall(ce *ast.CallExpr, rules []config.Rule, importMap map[string]string) []analysis.Diagnostic {
	var result []analysis.Diagnostic

	se, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	pkg, ok := se.X.(*ast.Ident)
	if !ok {
		return nil
	}

	for _, rule := range rules {
		if rule.Type == config.CallRule {
			match1 := rule.Name.Match(se.Sel.Name)

			match2 := rule.Package.Match(importMap[pkg.Name])

			if match1 && match2 {
				result = append(result, analysis.Diagnostic{
					Pos:     ce.Pos(),
					Message: rule.String(),
				})
			}
		}
	}

	return result
}
