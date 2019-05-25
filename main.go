package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"
)

func doTmpl(filename string, tmpl *template.Template, v interface{}) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	err = tmpl.Execute(w, v)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var (
		fat     = flag.Bool("f", false, "do not minimize DFA")
		dot     = flag.Bool("d", false, "generate dot file")
		nfa     = flag.Bool("n", false, "generate NFA (no lexer)")
		prefix  = flag.String("p", "yy", "prefix")
		outfile = flag.String("o", "lex.yy", "output base name")
	)
	flag.Parse()

	var p parser
	if flag.NArg() != 1 {
		fmt.Println("need exact 1 argument")
		os.Exit(1)
	}
	p.Filename = flag.Arg(0)
	fin, err := os.Open(p.Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fin.Close()
	p.Parse(fin)
	p.Prefix = *prefix
	if !*nfa {
		p.mach = p.mach.dfa()
		if !*fat {
			p.mach = p.mach.minimize()
		}
	}
	if *nfa || *dot {
		doTmpl(*outfile+".dot", dotTmpl, p)
	}
	if !*nfa {
		name := *outfile + ".go"
		doTmpl(name, lexTmpl, p)
		addimports(name, []string{"io", "unicode/utf8"})
	}
}
