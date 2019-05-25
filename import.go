package main

import (
	"bufio"
	"go/ast"
	"go/format"
	goparser "go/parser"
	"go/token"
	"log"
	"os"
	"strconv"
)

func addimports(filename string, imports []string) {
	ma := map[string]struct{}{}
	for _, name := range imports {
		ma[name] = struct{}{}
	}
	fset := token.NewFileSet()
	f, err := goparser.ParseFile(fset, filename, nil, goparser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	for _, imp := range f.Imports {
		name, _ := strconv.Unquote(imp.Path.Value)
		delete(ma, name)
	}
	dcl := &ast.GenDecl{Tok: token.IMPORT}
	for name := range ma {
		dcl.Specs = append(dcl.Specs, &ast.ImportSpec{
			Path: &ast.BasicLit{
				Value: strconv.Quote(name),
			},
		})
	}
	f.Decls = append([]ast.Decl{dcl}, f.Decls...)

	fout, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()
	w := bufio.NewWriter(fout)
	defer w.Flush()
	err = format.Node(w, fset, f)
	if err != nil {
		log.Fatal(err)
	}
}
