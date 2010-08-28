package passwd

import (
	"os"
	"fmt"
	"bytes"
	"io/ioutil"
	"regexp"
	"sync"
	"strconv"
	"github.com/droundy/goadmin/ago"
)

type Field int
const (
	Name Field = 1<<iota
	Passwd
	Uid
	Gid
	Comment
	Home
	Shell
	All Field = Passwd | Uid | Gid | Comment | Home | Shell
)

type User struct {
	Name, Passwd string
	Uid, Gid int
	Comment, Home, Shell string
}

var havepasswords = false
var passwd = make(map[string]User)
var created = make(map[string]User)
var once sync.Once
func Get() map[string]User {
	once.Do(func () {
		x, e := ioutil.ReadFile("/etc/passwd")
		if e == nil {
			pwent := "([^:\n]+):([^:]+):([0-9]+):([0-9]+):([^:]*):([^:]+):([^:]+)\n"
			matches := regexp.MustCompile(pwent).FindAllSubmatch(x, -1)
			for _,match := range matches {
				username := string(match[1])
				passwr := string(match[2])
				uid, e := strconv.Atoi(string(match[3]))
				if e != nil {
					fmt.Println("Error reading uid of", username, string(match[3]))
					continue
				}
				gid, e := strconv.Atoi(string(match[4]))
				if e != nil {
					fmt.Println("Error reading gid of", username, string(match[4]))
					continue
				}
				comment := string(match[5])
				home := string(match[6])
				shell := string(match[7])
				passwd[username] = User{username, passwr, uid, gid, comment, home, shell}
			}
			x, e = ioutil.ReadFile("/etc/shadow")
			if e == nil {
				pwent := "([^:\n]+):([^:]+):[^:]*:[^:]*:[^:]*:[^:]*:[^:]*:[^:]*:[^:]*\n"
				matches := regexp.MustCompile(pwent).FindAllSubmatch(x, -1)
				for _,match := range matches {
					username := string(match[1])
					passwr := string(match[2])
					if us, ok := passwd[username]; ok {
						us.Passwd = passwr // update the password!
						passwd[username] = us
					}
				}
				havepasswords = true
			} else {
				fmt.Fprintln(os.Stderr, "Unable to read /etc/shadow.")
			}
		}
	})
	return passwd
}

func (u User) GoSet(f Field) {
	if !havepasswords && f & Passwd != 0 {
		fmt.Fprintln(os.Stderr, "Refusing to set password for",u.Name,", since we can't read /etc/shadow!")
		f = f ^ Passwd // Don't set password if we can't read /etc/shadow!
	}
	if f == 0 { return } // Nothing to set!
	ago.Import("github.com/droundy/goadmin/passwd")
	code := "e = passwd.User{"+strconv.Quote(u.Name)+","
	if Passwd & f != 0 {
		code += strconv.Quote(u.Passwd) + ","
	} else {
		code += `"",`
	}
	code += fmt.Sprint(u.Uid, ",", u.Gid, ",")
	code += fmt.Sprint(strconv.Quote(u.Comment), ",", strconv.Quote(u.Home), ",", strconv.Quote(u.Shell))
	code += fmt.Sprint("}.Set(", f, ")")
	code += fmt.Sprint("\n\tif e != nil { return }")
	ago.Code(code)
}

func (u User) Set(f Field) os.Error {
	Get()
	didsomething := false
	old, present := passwd[u.Name]
	if present {
		n := old
		if (f & Uid != 0 && u.Uid != old.Uid) {
			fmt.Println("Setting Uid of", u.Name, "to", u.Uid, "from", old.Uid)
			didsomething = true
			n.Uid = u.Uid
		}
		if (f & Gid != 0 && u.Gid != old.Gid) {
			fmt.Println("Setting Gid of", u.Name, "to", u.Gid, "from", old.Gid)
			didsomething = true
			n.Gid = u.Gid
		}
		if (f & Passwd != 0 && u.Passwd != old.Passwd) {
			fmt.Println("Setting Passwd of", u.Name, "to", u.Passwd, "from", old.Passwd)
			didsomething = true
			n.Passwd = u.Passwd
		}
		if (f & Comment != 0 && u.Comment != old.Comment) {
			fmt.Println("Setting Comment of", u.Name, "to", u.Comment, "from", old.Comment)
			didsomething = true
			n.Comment = u.Comment
		}
		if (f & Home != 0 && u.Home != old.Home) {
			fmt.Println("Setting Home of", u.Name, "to", u.Home, "from", old.Home)
			didsomething = true
			n.Home = u.Home
		}
		if (f & Shell != 0 && u.Shell != old.Shell) {
			fmt.Println("Setting Shell of", u.Name, "to", u.Shell, "from", old.Shell)
			didsomething = true
			n.Shell = u.Shell
		}
		passwd[u.Name] = n
	} else {
		for _,x := range passwd {
			if x.Uid == u.Uid {
				fmt.Println("Uid conflict: ", x.Name, "already has Uid", u.Uid);
				return os.NewError("Uid already exists as "+x.Name+": "+fmt.Sprint(u.Uid))
			}
		}
		fmt.Println("Creating", u.Name)
		passwd[u.Name] = u
		created[u.Name] = u
		didsomething = true
	}
	if didsomething {
		update()
	}
	return nil
}

func update() os.Error {
	Get()
	x, e := ioutil.ReadFile("/etc/passwd")
	if e != nil { return e }

  pwdfile,e := os.Open("/etc/passwd.new", os.O_WRONLY + os.O_TRUNC + os.O_CREAT, 0644)
	if e != nil { return e }
	defer pwdfile.Close()
	pwent := regexp.MustCompile("^([^:\n]+):([^:]+):([0-9]+):([0-9]+):([^:]*):([^:]+):([^:]+)\n$")
	for _,l := range bytes.SplitAfter(x, []byte("\n"), -1) {
		matches := pwent.FindSubmatch(l)
		if matches != nil {
			u, ok := passwd[string(matches[1])]
			if ok {
				_,e = fmt.Fprintf(pwdfile, "%s:x:%d:%d:%s:%s:%s\n", u.Name, u.Uid, u.Gid, u.Comment, u.Home, u.Shell)
				if e != nil { return e }
			} else {
				_,e = pwdfile.Write(l)
				if e != nil { return e }
			}
		} else {
			_,e = pwdfile.Write(l)
			if e != nil { return e }
		}
	}
	for _, u := range created {
		fmt.Fprintf(pwdfile, "%s:x:%d:%d:%s:%s:%s\n", u.Name, u.Uid, u.Gid, u.Comment, u.Home, u.Shell)
		if e != nil { return e }
	}
	pwdfile.Close()

	x, e = ioutil.ReadFile("/etc/shadow")
	if e != nil { return e }
  shadow,e := os.Open("/etc/shadow.new", os.O_WRONLY + os.O_TRUNC + os.O_CREAT, 0600)
	if e != nil { return e }
	defer shadow.Close()

	pwent = regexp.MustCompile("([^:\n]+):([^:]+):([^:]*:[^:]*:[^:]*:[^:]*:[^:]*:[^:]*:[^:]*)\n")
	for _,l := range bytes.SplitAfter(x, []byte("\n"), -1) {
		matches := pwent.FindSubmatch(l)
		if matches != nil {
			u, ok := passwd[string(matches[1])]
			if ok {
				_,e = fmt.Fprintf(shadow, "%s:%s:%s\n", u.Name, u.Passwd, matches[3])
				if e != nil { return e }
			} else {
				_,e = shadow.Write(l)
				if e != nil { return e }
			}
		} else {
			_,e = shadow.Write(l)
			if e != nil { return e }
		}
	}
	for _, u := range created {
		_,e = fmt.Fprintf(shadow, "%s:%s:::::::\n", u.Name, u.Passwd)
		if e != nil { return e }
	}
	shadow.Close()

	e = os.Rename("/etc/passwd.new", "/etc/passwd")
	if e != nil { return e }
	e = os.Rename("/etc/shadow.new", "/etc/shadow")
	if e != nil { return e }
	return nil
}
