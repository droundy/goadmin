package ago

import (
	"os"
	"fmt"
	"strconv"
	"sync"
	"github.com/droundy/goadmin/deps"
)

func panicon(err os.Error) {
	if err != nil {
		panic(err)
	}
}

func archnum() string {
	switch os.Getenv("GOARCH") {
	case "386": return "8"
	case "amd64": return "6"
		// what was the other one called?
	}
	return "5"
}

func Compile(outname string, files []string) (e os.Error) {
	if len(files) < 1 {
		return os.NewError("go.Compile requires at least one file argument.");
	}
	objname := "_go_."+archnum()

	args := make([]string, len(files) + 2)
	args[0] = "-o"
	args[1] = objname
	for i,f := range files {
		args[i+2] = f
	}
	e = deps.Execs(archnum()+"g", args)
	if e != nil { return }
	return deps.Exec(archnum()+"l", "-o", outname, objname)
}

var createcodeonce sync.Once
var imports = make(chan string)
var code = make(chan string)
var done = make(chan string)
var errr = make(chan os.Error)

func init() {
	go func () {
		imps := make(map[string]bool)
		imps["github.com/droundy/goadmin/deps"] = true
		allcode := ""
		for {
			select {
			case i := <- imports:
				imps[i] = true
			case c := <- code:
				allcode = allcode + "\n\t" + c
			case varname := <- done:
				_,e := fmt.Print(`package main

import (
`)
				if e != nil {
					fmt.Println(e)
					errr <- e
					return
				}
				for i := range imps {
					_, e = fmt.Print("\t", strconv.Quote(i), "\n")
					if e != nil {
						fmt.Println(e)
						errr <- e
						return
					}
				}
				_, e = fmt.Print(`)

var `, varname, ` = deps.Run(func () (e deps.Error) {`)
				if e != nil {
					fmt.Println(e)
					errr <- e
					return
				}
				_, e = fmt.Println(allcode)
				if e != nil {
					fmt.Println(e)
					errr <- e
					return
				}
				_, e = fmt.Println("\treturn\n})")
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

func Code(c string) {
	code <- c
}

func Print(varname string) os.Error {
	done <- varname
	return <- errr
}
