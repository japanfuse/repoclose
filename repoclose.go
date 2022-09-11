package repoclose

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

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

func initialize(pass *analysis.Pass) {
	gPass = pass
}

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	initialize(pass)

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
				fmt.Println(ident)
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
	ident := exprToString(stmt.Lhs[0])

	ident = strings.ToLower(ident)
	return strings.HasSuffix(ident, "repo")
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
