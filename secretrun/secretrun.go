package main

import (
	"fmt"
	"os"
	"syscall"
	"github.com/droundy/goadmin/ago/compile"
	"github.com/droundy/goadmin/deps"
	"github.com/droundy/goopt"
)

var outname = goopt.String([]string{"-o"}, "FILENAME.go", "the name of the generated go file")

func main() {
	goopt.Parse(func() []string { return []string{} })

	syscall.Umask(0077) // Turn off read/write/execute priviledges for others

	if len(goopt.Args) < 1 {
		fmt.Println("You need to provide a go file argument.")
		os.Exit(1);
	}

	execname := goopt.Args[0]+".secret"
	e := compile.Compile(execname, goopt.Args)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	if *outname == "FILENAME.go" {
		fmt.Fprintln(os.Stderr, "secretrun requires a -o argument!")
		os.Exit(1)
	}
	e = deps.Exec("./"+execname, *outname)
	os.Remove(execname)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
