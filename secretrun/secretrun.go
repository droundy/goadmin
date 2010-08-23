package main

import (
	"fmt"
	"os"
	"syscall"
	"github.com/droundy/goadmin/ago"
	"github.com/droundy/goadmin/deps"
	"github.com/droundy/goopt"
)

func main() {
	goopt.Parse(func() []string { return []string{} })

	syscall.Umask(0077) // Turn off read/write/execute priviledges for others

	if len(goopt.Args) < 1 {
		fmt.Println("You need to provide a go file argument.")
		os.Exit(1);
	}

	execname := goopt.Args[0]+".secret"
	e := ago.Compile(execname, goopt.Args)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	e = deps.Exec("./"+execname)
	os.Remove(execname)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
