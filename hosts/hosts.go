package hosts

import (
	"fmt"
	"os"
	"strings"
	"io/ioutil"
	"regexp"
	"sync"
	"strconv"
	"github.com/droundy/goadmin/gomakefile"
	"github.com/droundy/goadmin/ago"
)

type Entry struct {
	IpAddress string
	CanonicalName string
	Aliases []string
}

var myhosts = make(map[string]Entry)
var byip = make(map[string]Entry)
var once sync.Once
func Get() map[string]Entry {
	once.Do(func () {
		x, e := ioutil.ReadFile("/etc/hosts")
		if e == nil {
			gomakefile.AddDep("/etc/hosts")
			matchhosts := regexp.MustCompile("(^|\n)([^#\n\t ]+)[\t ]+([^\n\t ]+)[\t ]*([^\n]*)\n")
			matchaliases := regexp.MustCompile("[^\t ]+")
			matches := matchhosts.FindAllSubmatch(x, -1)
			for _,match := range matches {
				as := matchaliases.FindAllString(string(match[4]),-1)
				ent := Entry{ string(match[2]), string(match[3]), as }

				byip[ent.IpAddress] = ent
				myhosts[ent.CanonicalName] = ent
				myhosts[ent.IpAddress] = ent
				for _,a := range ent.Aliases {
					myhosts[a] = ent
				}
			}
		}
	})
	return myhosts
}

func (ent Entry) GoSet() {
	ago.Import("github.com/droundy/goadmin/hosts")
	code := "e = hosts.Entry{"+strconv.Quote(ent.IpAddress)+","+
		strconv.Quote(ent.CanonicalName)+", []string{"
	as := make([]string, len(ent.Aliases))
	for i := range ent.Aliases {
		as[i] = strconv.Quote(ent.Aliases[i])
	}
	code += strings.Join(as, ", ") +  " } }.Set()"
	ago.Code(code)
}

func (ent Entry) Set() os.Error {
	Get()
	byip[ent.IpAddress] = ent
	myhosts[ent.CanonicalName] = ent
	for _, a := range ent.Aliases {
		myhosts[a] = ent
	}
	update()
	return nil
}

func update() os.Error {
	Get()
	x, e := ioutil.ReadFile("/etc/hosts")
	if e != nil { return e }
	matchhost := regexp.MustCompile("([^\n\t ]+)[\t ]+([^\n\t ]+)[\t ]*([^\n]*)([\t ]+[^\t \n]+)*[\t ]*\n")

	e = os.Remove("/etc/hosts.new")
	_, ispatherror := e.(*os.PathError) // *os.PathError means file doesn't exist (which is okay)
	if e != nil && !ispatherror { return e }

  hfile,e := os.Open("/etc/hosts.new", os.O_WRONLY + os.O_TRUNC + os.O_CREAT + os.O_EXCL, 0644)
	if e != nil { return e }
	defer hfile.Close()

	alreadydone := make(map[string]bool)
	for _, line := range strings.SplitAfter(string(x), "\n", -1) {
		ms := matchhost.FindStringSubmatch(line)
		if len(line) == 0 || line[0] == '#' || len(ms) < 3 {
			hfile.Write([]byte(line))
		} else {
			// We have an entry!
			ip := ms[1]
			ent, ok := byip[ip]
			if ok {
				// This is one that we've already to got loaded...
				fmt.Fprint(hfile, ip, "\t", ent.CanonicalName)
				for _,a := range ent.Aliases {
					fmt.Fprint(hfile, "\t", a)
				}
				fmt.Fprint(hfile, "\n")
				alreadydone[ip] = true
			} else {
				// It's present in the file, but not in our map, so we need to
				// see if it conflicts with one that's in the map...
				isok := true
				for _,a := range ms[2:] {
					_, confl := myhosts[string(a)]
					isok = isok && !confl
				}
				if isok {
					// It doesn't conflict with anything we've got, so let's preserve it.
					hfile.Write([]byte(line))
				} else {
					// I'm being a little paranoid here, not wanting to remove
					// things that a user might have added.
					fmt.Fprint(hfile, "#")
					hfile.Write([]byte(line[0:len(line)-1]))
					fmt.Fprintln(hfile, " # commented because it conflicts")
				}
			}
		}
	}
	for _,ent := range byip {
		_, alreadydone := alreadydone[ent.IpAddress]
		if !alreadydone {
			fmt.Fprint(hfile, ent.IpAddress, "\t", ent.CanonicalName)
			for _,a := range ent.Aliases {
				fmt.Fprint(hfile, "\t", a)
			}
			fmt.Fprint(hfile, "\n")
		}
	}

	hfile.Close()
	return os.Rename("/etc/hosts.new", "/etc/hosts")
}
