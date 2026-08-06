package check

import (
	"fmt"
	"go/ast"
	"strings"

	iradix "github.com/hashicorp/go-immutable-radix"
	"github.com/ormanli/mixingcheck/internal/config"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Version defines version.
var Version = "development"

func initializeTree(c config.Packages) (*iradix.Tree, error) {
	r := iradix.New()

	for name, pkg := range c {
		for i := range pkg.Rules {
			if pkg.Rules[i].Type != config.StructRule && pkg.Rules[i].Type != config.CallRule {
				return nil, fmt.Errorf("rule %d in %s has invalid type", i, name)
			}

			if pkg.Rules[i].Package.Value == "" {
				return nil, fmt.Errorf("rule %d in %s has empty package", i, name)
			}

			err := pkg.Rules[i].Package.Compile(pkg.Rules[i].Package.Value)
			if err != nil {
				return nil, fmt.Errorf("rule %d in %s has invalid regex, %w", i, name, err)
			}

			if pkg.Rules[i].Name.Value == "" {
				return nil, fmt.Errorf("rule %d in %s has empty name", i, name)
			}

			err = pkg.Rules[i].Name.Compile(pkg.Rules[i].Name.Value)
			if err != nil {
				return nil, fmt.Errorf("rule %d in %s has invalid regex, %w", i, name, err)
			}

			pkg.Rules[i].Message = pkg.Rules[i].String()
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

	runner := &runner{tree: tree}

	return &analysis.Analyzer{
		Name:     "mixingcheck",
		Doc:      fmt.Sprintf("check forbidden structs and method calls. Version %s", Version),
		Run:      runner.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}, nil
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

func (r *runner) run(pass *analysis.Pass) (any, error) {
	rules := r.gatherRules(pass.Pkg.Path())
	if len(rules) == 0 {
		return nil, nil
	}

	var structRules, callRules []config.Rule
	for _, rule := range rules {
		switch rule.Type {
		case config.StructRule:
			structRules = append(structRules, rule)
		case config.CallRule:
			callRules = append(callRules, rule)
		}
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.File)(nil),
		(*ast.SelectorExpr)(nil),
		(*ast.CallExpr)(nil),
	}

	var importMap map[string]string

	insp.Preorder(nodeFilter, func(n ast.Node) {
		switch node := n.(type) {
		case *ast.File:
			importMap = r.extractImports(node)
		case *ast.SelectorExpr:
			for _, diagnostic := range checkSelector(node, structRules, importMap) {
				pass.Report(diagnostic)
			}
		case *ast.CallExpr:
			for _, diagnostic := range checkCall(node, callRules, importMap) {
				pass.Report(diagnostic)
			}
		}
	})

	return nil, nil
}

func (r *runner) gatherRules(packageName string) []config.Rule {
	packageNameAsBytes := []byte(packageName)

	var baseRules []config.Rule
	if value, exist := r.tree.Root().Get(packageNameAsBytes); exist {
		pkg := value.(config.Package)

		if pkg.IgnoreParentRules {
			return pkg.Rules
		}

		baseRules = pkg.Rules
	}

	var otherRules []config.Rule

	r.tree.Root().WalkPath(packageNameAsBytes, func(k []byte, v any) bool {
		// k is always a byte-prefix of packageNameAsBytes (guaranteed by WalkPath),
		// but a byte-prefix isn't necessarily an ancestor package: "a/c" is a
		// byte-prefix of "a/cd" without being its parent, so only accept exact
		// matches (handled above via Get) or matches that end on a "/" boundary.
		if len(k) == len(packageNameAsBytes) {
			return false
		}

		if packageNameAsBytes[len(k)] != '/' {
			return false
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

	rules := make([]config.Rule, 0, len(baseRules)+len(otherRules))
	rules = append(rules, baseRules...)
	rules = append(rules, otherRules...)

	return rules
}

func checkSelector(se *ast.SelectorExpr, rules []config.Rule, importMap map[string]string) []analysis.Diagnostic {
	pkg, ok := se.X.(*ast.Ident)
	if !ok {
		return nil
	}

	importPath, ok := importMap[pkg.Name]
	if !ok {
		return nil
	}

	var result []analysis.Diagnostic

	for _, rule := range rules {
		if rule.Name.Match(se.Sel.Name) && rule.Package.Match(importPath) {
			result = append(result, analysis.Diagnostic{
				Pos:     se.Pos(),
				Message: rule.Message,
			})
		}
	}

	return result
}

func checkCall(ce *ast.CallExpr, rules []config.Rule, importMap map[string]string) []analysis.Diagnostic {
	se, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	pkg, ok := se.X.(*ast.Ident)
	if !ok {
		return nil
	}

	importPath, ok := importMap[pkg.Name]
	if !ok {
		return nil
	}

	var result []analysis.Diagnostic

	for _, rule := range rules {
		if rule.Name.Match(se.Sel.Name) && rule.Package.Match(importPath) {
			result = append(result, analysis.Diagnostic{
				Pos:     ce.Pos(),
				Message: rule.Message,
			})
		}
	}

	return result
}
