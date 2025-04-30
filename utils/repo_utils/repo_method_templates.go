package repo_utils

import (
	"github.com/jinzhu/inflection"
	"go/ast"
	"go/token"
	"goblin/cli_config"
	"goblin/utils"
)

//////// bodies

var generateCreateMethodBody = func(modelPascalCase string) []ast.Stmt {
	return []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("err")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("r.db"),
							Sel: ast.NewIdent("Create"),
						},
						Args: []ast.Expr{ast.NewIdent(utils.PascalToCamel(modelPascalCase))},
					},
					Sel: ast.NewIdent("Error"),
				},
			},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("nil"),
							ast.NewIdent("err"),
						},
					},
				},
			},
		},
		&ast.ReturnStmt{
			Results: []ast.Expr{
				ast.NewIdent(utils.PascalToCamel(modelPascalCase)),
				ast.NewIdent("nil"),
			},
		},
	}
}

var generateUpdateMethodBody = func(modelPascalCase string) []ast.Stmt {
	return []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("err")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("r.db"),
							Sel: ast.NewIdent("Save"),
						},
						Args: []ast.Expr{ast.NewIdent(utils.PascalToCamel(modelPascalCase))},
					},
					Sel: ast.NewIdent("Error"),
				},
			},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("nil"),
							ast.NewIdent("err"),
						},
					},
				},
			},
		},
		&ast.ReturnStmt{
			Results: []ast.Expr{
				ast.NewIdent(utils.PascalToCamel(modelPascalCase)),
				ast.NewIdent("nil"),
			},
		},
	}
}

var generateDeleteMethodBody = func(modelPascalCase string) []ast.Stmt {
	return []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("err")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("r.db"),
							Sel: ast.NewIdent("Delete"),
						},
						Args: []ast.Expr{ast.NewIdent(utils.PascalToCamel(modelPascalCase))},
					},
					Sel: ast.NewIdent("Error"),
				},
			},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("err"),
						},
					},
				},
			},
		},
		&ast.ReturnStmt{
			Results: []ast.Expr{
				ast.NewIdent("nil"),
			},
		},
	}
}

var generateListAllMethodBody = func(modelPascalCase string) []ast.Stmt {
	modelsPackage := cli_config.CliConfig.ModelsFolderPath + "."
	return []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{ast.NewIdent(inflection.Plural(utils.PascalToCamel(modelPascalCase)))},
						Type: &ast.ArrayType{
							Elt: ast.NewIdent(modelsPackage + modelPascalCase),
						},
					},
				},
			},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("err")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("r.db"),
							Sel: ast.NewIdent("Find"),
						},
						Args: []ast.Expr{&ast.UnaryExpr{
							Op: token.AND,
							X:  ast.NewIdent(inflection.Plural(utils.PascalToCamel(modelPascalCase))),
						}},
					},
					Sel: ast.NewIdent("Error"),
				},
			},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("nil"),
							ast.NewIdent("err"),
						},
					},
				},
			},
		},
		&ast.ReturnStmt{
			Results: []ast.Expr{
				ast.NewIdent(inflection.Plural(utils.PascalToCamel(modelPascalCase))),
				ast.NewIdent("nil"),
			},
		},
	}
}

var generateListWithPaginationMethodBody = func(model string) []ast.Stmt {
	return []ast.Stmt{
		&ast.ReturnStmt{
			Results: []ast.Expr{
				ast.NewIdent("nil"),
				ast.NewIdent("nil"),
			},
		},
	}
}

var generateGetByUuidMethodBody = func(modelPascalCase string) []ast.Stmt {
	modelsPackage := cli_config.CliConfig.ModelsFolderPath + "."
	return []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{ast.NewIdent(utils.PascalToCamel(modelPascalCase))},
						Type:  ast.NewIdent("*" + modelsPackage + modelPascalCase),
					},
				},
			},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("err")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent("r.db"),
									Sel: ast.NewIdent("Where"),
								},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: `"uuid = ?"`},
									ast.NewIdent("uuid"),
								},
							},
							Sel: ast.NewIdent("First"),
						},
						Args: []ast.Expr{
							ast.NewIdent(utils.PascalToCamel(modelPascalCase)),
						},
					},
					Sel: ast.NewIdent("Error"),
				},
			},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("nil"),
							ast.NewIdent("err"),
						},
					},
				},
			},
		},
		&ast.ReturnStmt{
			Results: []ast.Expr{
				ast.NewIdent(utils.PascalToCamel(modelPascalCase)),
				ast.NewIdent("nil"),
			},
		},
	}
}

func generateMethodBody(repoMethod Method, modelPascalCase string) []ast.Stmt {

	switch repoMethod {
	case Create:
		return generateCreateMethodBody(modelPascalCase)
	case Update:
		return generateUpdateMethodBody(modelPascalCase)
	case Delete:
		return generateDeleteMethodBody(modelPascalCase)
	case ListAll:
		return generateListAllMethodBody(modelPascalCase)
	case ListWithPagination:
		return generateListWithPaginationMethodBody(utils.PascalToCamel(modelPascalCase))
	case GetByUuid:
		return generateGetByUuidMethodBody(modelPascalCase)
	default:
		return []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					ast.NewIdent(utils.PascalToCamel(modelPascalCase)),
					ast.NewIdent("nil"),
				},
			},
		}
	}
}

//////// params

var generateCreateMethodParams = func(modelPascalCase, modelType string) []*ast.Field {
	return []*ast.Field{
		{
			Names: []*ast.Ident{ast.NewIdent(utils.PascalToCamel(modelPascalCase))},
			Type:  ast.NewIdent("*" + modelType),
		},
	}
}

var generateUpdateMethodParams = func(modelPascalCase, modelType string) []*ast.Field {
	return []*ast.Field{
		{
			Names: []*ast.Ident{ast.NewIdent(utils.PascalToCamel(modelPascalCase))},
			Type:  ast.NewIdent("*" + modelType),
		},
	}
}

var generateDeleteMethodParams = func(modelPascalCase, modelType string) []*ast.Field {
	return []*ast.Field{
		{
			Names: []*ast.Ident{ast.NewIdent(utils.PascalToCamel(modelPascalCase))},
			Type:  ast.NewIdent("*" + modelType),
		},
	}
}

var generateListAllMethodParams = func() []*ast.Field {
	return []*ast.Field{}
}

var generateListWithPaginationMethodParams = func() []*ast.Field {
	return []*ast.Field{}
}

var generateGetByUuidMethodParams = func() []*ast.Field {
	return []*ast.Field{
		{
			Names: []*ast.Ident{ast.NewIdent("uuid")},
			Type:  ast.NewIdent("string"),
		},
	}
}

func generateMethodParams(repoMethod Method, modelPascalCase, modelDataType string) []*ast.Field {
	switch repoMethod {
	case Create:
		return generateCreateMethodParams(modelPascalCase, modelDataType)
	case Update:
		return generateUpdateMethodParams(modelPascalCase, modelDataType)
	case Delete:
		return generateDeleteMethodParams(modelPascalCase, modelDataType)
	case ListAll:
		return generateListAllMethodParams()
	case ListWithPagination:
		return generateListWithPaginationMethodParams()
	case GetByUuid:
		return generateGetByUuidMethodParams()
	default:
		return []*ast.Field{}
	}
}

///////// results

var generateCreateMethodResults = func(modelPascalCase, modelType string) []*ast.Field {
	return []*ast.Field{
		{Type: ast.NewIdent("*" + modelType)},
		{Type: ast.NewIdent("error")},
	}
}

var generateUpdateMethodResults = func(modelPascalCase, modelType string) []*ast.Field {
	return []*ast.Field{
		{Type: ast.NewIdent("*" + modelType)},
		{Type: ast.NewIdent("error")},
	}
}

var generateDeleteMethodResults = func() []*ast.Field {
	return []*ast.Field{
		{Type: ast.NewIdent("error")},
	}
}

var generateListAllMethodResults = func(modelPascalCase, modelType string) []*ast.Field {
	return []*ast.Field{
		{
			Type: &ast.ArrayType{
				Elt: ast.NewIdent(modelType),
			},
		},
		{
			Type: ast.NewIdent("error"),
		},
	}
}

var generateListWithPaginationMethodResults = func(modelPascalCase, modelType string) []*ast.Field {
	return []*ast.Field{
		{
			Type: &ast.ArrayType{
				Elt: ast.NewIdent(modelType),
			},
		},
		{
			Type: ast.NewIdent("error"),
		},
	}
}

var generateGetByUuidMethodResults = func(modelPascalCase, modelType string) []*ast.Field {
	return []*ast.Field{
		{Type: ast.NewIdent("*" + modelType)},
		{Type: ast.NewIdent("error")},
	}
}

func generateMethodResults(repoMethod Method, modelPascalCase, modelDataType string) []*ast.Field {
	switch repoMethod {
	case Create:
		return generateCreateMethodResults(modelPascalCase, modelDataType)
	case Update:
		return generateUpdateMethodResults(modelPascalCase, modelDataType)
	case Delete:
		return generateDeleteMethodResults()
	case ListAll:
		return generateListAllMethodResults(modelPascalCase, modelDataType)
	case ListWithPagination:
		return generateListWithPaginationMethodResults(modelPascalCase, modelDataType)
	case GetByUuid:
		return generateGetByUuidMethodResults(modelPascalCase, modelDataType)
	default:
		return []*ast.Field{}
	}
}
