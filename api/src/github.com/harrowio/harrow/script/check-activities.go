package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
)

func main() {
	files := token.NewFileSet()
	packages, err := parser.ParseDir(files, "activities", nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	found := map[string]ast.Stmt{}
	pkg := packages["activities"]
	for _, file := range pkg.Files {
		handleFile(found, file)
	}

	hasUnregisteredActivities := false
	for activity, registeredBy := range found {
		if registeredBy == nil {
			fmt.Fprintf(os.Stdout, "unregistered activity payload: %s\n", activity)
			hasUnregisteredActivities = true
		}
	}

	if hasUnregisteredActivities {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func handleFile(found map[string]ast.Stmt, file *ast.File) {
	for _, decl := range file.Decls {
		handleDecl(found, file, decl)
	}
}

func handleDecl(found map[string]ast.Stmt, file *ast.File, decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Name.Name == "init" {
			handleInit(found, d)
		} else {
			_, exists := found[d.Name.Name]
			if returnsActivity(d) && !exists {
				found[d.Name.Name] = nil
			}
		}
	}
}

func handleInit(found map[string]ast.Stmt, f *ast.FuncDecl) {
	for _, stmt := range f.Body.List {
		expr, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}

		call, ok := expr.X.(*ast.CallExpr)
		if !ok {
			continue
		}

		fun, ok := call.Fun.(*ast.Ident)
		if !ok {
			continue
		}

		if fun.Name != "registerPayload" {
			continue
		}

		arg, ok := call.Args[0].(*ast.CallExpr)
		if !ok {
			continue
		}

		activityConstructor, ok := arg.Fun.(*ast.Ident)
		if !ok {
			continue
		}

		found[activityConstructor.Name] = stmt
	}
}

func returnsActivity(f *ast.FuncDecl) bool {
	if f.Type.Results == nil || len(f.Type.Results.List) != 1 {
		return false
	}

	returns := f.Type.Results.List
	returnValue := returns[0].Type
	starExpr, ok := returnValue.(*ast.StarExpr)
	if !ok {
		return false
	}

	selector, ok := starExpr.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkgname, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	typename := selector.Sel

	return fmt.Sprintf("%s.%s", pkgname, typename) == "domain.Activity"
}
