package repoclose

import (
	"go/ast"
	"go/types"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "repoclose is ..."

const (
	Controller = "*Controller"
	Repository = "portxserver/domain/repository.Transaction"
)

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "repoclose",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

var RepositoryInterface *types.Interface

var gPass *analysis.Pass

var allPkgs []*types.Package
var set mapset.Set[*types.Package]
var iTransaction *types.Interface

func initialize(pass *analysis.Pass) {
	gPass = pass

	allPkgs = pass.Pkg.Imports()
	set = mapset.NewSet(allPkgs...)
	for _, p := range gPass.Pkg.Imports() {
		setImportedPkgs(p)
	}
}

func setImportedPkgs(pkg *types.Package) {
	for _, p := range pkg.Imports() {
		if !set.Contains(p) {
			set.Add(p)
			allPkgs = append(allPkgs, p)
			setImportedPkgs(p)
		}
	}
}

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	initialize(pass)

	obj := analysisutil.ObjectOf(gPass, "portxserver/domain/repository", "Transaction")
	if obj == nil {
		return nil, nil
	}

	var ok bool
	iTransaction, ok = obj.Type().(*types.Interface)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.FuncDecl:
			// TODO: Controller のメソッドのみに限定したい
			check(n.Body.List)
		}
	})

	return nil, nil
}

func check(stmts []ast.Stmt) {
	repoAssigns := map[string]ast.Stmt{}
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *ast.AssignStmt:
			if isRepoAssignment(stmt) {
				lhsName := exprToString(stmt.Lhs[0])
				repoAssigns[lhsName] = stmt
			}
		}
	}

	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *ast.DeferStmt:
			if isCloseCall(stmt.Call) {
				ident := getReceiverIdent(stmt.Call.Fun)
				delete(repoAssigns, ident)
			}
		}
	}

	for _, v := range repoAssigns {
		gPass.Reportf(v.Pos(), "this repository is not closed")
	}
}

// 左辺の識別子が repo という suffix を持っていたら true
func isRepoAssignment(stmt *ast.AssignStmt) bool {
	t := gPass.TypesInfo.TypeOf(stmt.Lhs[0])
	return types.Implements(t, iTransaction)
}

func isCloseCall(stmt *ast.CallExpr) bool {
	ident := exprToString(stmt.Fun)
	return strings.HasSuffix(ident, "Rollback")
}

func exprToString(expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.SelectorExpr:
		return exprToString(expr.X) + expr.Sel.Name
	default:
		return ""
	}
}

func getReceiverIdent(expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.SelectorExpr:
		return exprToString(expr.X)
	default:
		return ""
	}
}
