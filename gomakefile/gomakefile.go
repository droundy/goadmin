package gomakefile

import (
	"os"
	"io/ioutil"
	"fmt"
	"regexp"
	"strconv"
	"github.com/droundy/goopt"
	"github.com/droundy/goadmin/ago"
)

var mymkfile = goopt.String([]string{"--makefile"}, "Makefile", "name of makefile to update")
var mytarget = goopt.String([]string{"--target"}, "TARGET", "name of target we're building")

func SetMakefile(mkfile string) {
	mymkfile = &mkfile
}

func SetTarget(targ string) {
	mytarget = &targ
}

func GoAddDep(dependency string) {
	ago.Import("github.com/droundy/goadmin/gomakefile")
	ago.Declare("var changed_files = make(map[string]bool)")
	ago.Declare("var this_file_changed = false")
	code := "this_file_changed, e = gomakefile.AddDep("+strconv.Quote(dependency)+")"
	code += fmt.Sprint("\n\tif this_file_changed { changed_files[`Makefile`] = true }")
	ago.Code(code)
}

func AddDep(dependency string) (needswriting bool, e os.Error) {
	mkfile := *mymkfile
	target := *mytarget
	if target == "TARGET" {
		if ago.OutputName != "" {
			target = ago.OutputName
	} else {
			return
		}
	}
	oldf, e := ioutil.ReadFile(mkfile)
	if e != nil {
		needswriting = true
		myf, e := os.Open(mkfile, os.O_WRONLY + os.O_TRUNC + os.O_CREAT + os.O_EXCL, 0666)
		if e != nil {
			return needswriting, e
		}
		defer myf.Close()
		_, e = fmt.Fprintln(myf, "\n", target, ": ", dependency)
		return needswriting, e
	} else {
		// The following regular expression fails to parse certain valid makefiles.
		// In particular, the target must not show up on the first line of the file, and
		// each dependency must be preceded by a space or tab.
		amok := regexp.MustCompile("\n"+regexp.QuoteMeta(target)+":([^\n]|\\\\\n|\\\\\r\n)*[ \t]"+
			regexp.QuoteMeta(dependency) + "([ \t\r\n#]|$)")
		if amok.Find(oldf) != nil {
			// The dependency is already in place...
			return
		}
		needswriting = true
		havetarget := regexp.MustCompile("\n"+regexp.QuoteMeta(target)+":([^\n]|\\\\\n|\\\\\r\n)*\n")
		if ind := havetarget.FindIndex(oldf); ind != nil {
			// We don't need to truncate it, since it's only going to get longer!
			myf, e := os.Open(mkfile+".new", os.O_WRONLY + os.O_CREAT, 0666)
			if e != nil {
				return needswriting, e
			}
			defer myf.Close()
			start := ind[0]
			end := ind[1]
			_,e = myf.Write(oldf[0:start])
			if e != nil {
				return needswriting, e
			}
			_,e = myf.Write(oldf[start:end-1])
			if e != nil {
				return needswriting, e
			}
			_,e = myf.Write([]byte("\\\n\t"+dependency))
			if e != nil {
				return needswriting, e
			}
			_,e = myf.Write(oldf[end-1:])
			return needswriting, e
		} else {
			newf := ""
			if oldf[len(oldf)-1] == '\n' {
				newf = string(oldf) + target + ": " + dependency + "\n"
			} else {
				newf = string(oldf) + "\n" + target + ": " + dependency + "\n"
			}
			e = ioutil.WriteFile(mkfile, []byte(newf), 0666)
			return needswriting, e
		}
	}
	return
}
