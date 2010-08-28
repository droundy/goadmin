package ago

import (
	"os"
	"fmt"
	"strconv"
	"sync"
)

var OutputName = ""

var createcodeonce sync.Once
var imports = make(chan string)
var code = make(chan string)
var declare = make(chan string)
var done = make(chan string)
var errr = make(chan os.Error)

func init() {
	if len(os.Args) == 2 {
		OutputName = os.Args[1]
	}
	go func () {
		imps := make(map[string]bool)
		imps["github.com/droundy/goadmin/deps"] = true
		allcode := ""
		declarations := make(map[string]bool)
		for {
			select {
			case i := <- imports:
				imps[i] = true
			case c := <- code:
				allcode = allcode + "\n\t" + c
			case d := <- declare:
				declarations[d] = true
			case varname := <- done:
				if len(OutputName) == 0 {
					fmt.Println("Need an output name argument!")
					errr <- os.NewError("Need an output name argument!")
					return
				}
				os.Remove(OutputName)
				outf,e := os.Open(OutputName, os.O_WRONLY + os.O_TRUNC + os.O_CREAT + os.O_EXCL, 0600)
				if e != nil {
					fmt.Println(e)
					errr <- e
					return
				}
				_,e = fmt.Fprint(outf, `package main

import (
`)
				if e != nil {
					fmt.Println(e)
					errr <- e
					return
				}
				for i := range imps {
					_, e = fmt.Fprint(outf, "\t", strconv.Quote(i), "\n")
					if e != nil {
						fmt.Println(e)
						errr <- e
						return
					}
				}
				_, e = fmt.Fprint(outf, `)

var `, varname, ` = deps.Run(func () (e deps.Error) {`)
				if e != nil {
					fmt.Println(e)
					errr <- e
					return
				}
				for d := range declarations {
					_, e = fmt.Fprint(outf, "\t", d, "\n")
					if e != nil {
						fmt.Println(e)
						errr <- e
						return
					}
				}
				_, e = fmt.Fprintln(outf, allcode)
				if e != nil {
					fmt.Println(e)
					errr <- e
					return
				}
				_, e = fmt.Fprintln(outf, "\treturn\n})")
				if e != nil {
					fmt.Println(e)
					errr <- e
					return
				}
				errr <- e
				return
			}
		}
	}()
}

func Import(i string) {
	imports <- i
}

func Declare(c string) {
	declare <- c
}

func Code(c string) {
	code <- c
}

func Print(varname string) os.Error {
	done <- varname
	return <- errr
}
