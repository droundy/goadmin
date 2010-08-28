package file

import (
	"os"
	"io"
	"io/ioutil"
	"fmt"
	"syscall"
	"strconv"
	"github.com/droundy/goadmin/ago"
)

type Field int
const (
	Name Field = 1<<iota
	Contents
	Uid
	Gid
	Perms
	All Field = Contents | Uid | Gid | Perms
)

type Data struct {
	Name, Contents string
	Uid, Gid int
	Perms uint32
}

func (f Data) Set(which Field) (needswriting bool, e os.Error) {
	stat, e := os.Lstat(f.Name)
	if e == nil {
		switch stat.Mode & syscall.S_IFMT {
		case syscall.S_IFLNK:
			e = os.Remove(f.Name)
			if e != nil { return false, e }
		case syscall.S_IFREG:
			// Just an ordinary file.
			if Perms & which != 0 && stat.Permission() != f.Perms {
				e = os.Chmod(f.Name, f.Perms)
				if e != nil { return false, e }
			} else {
				f.Perms = stat.Permission()
			}
			// Set the ownership if needed...
			if Gid & which == 0 { f.Gid = stat.Gid }
			if Uid & which == 0 { f.Uid = stat.Uid }
			e = os.Lchown(f.Name, f.Uid, f.Gid)
			if e != nil { return false, e }
		case syscall.S_IFDIR:
			return false, os.NewError(f.Name+" is a directory, not a file!")
		}
		e = os.Remove(f.Name+".new")
		_, ispatherror := e.(*os.PathError) // *os.PathError means .new file doesn't exist (which is okay)
		if e != nil && !ispatherror { return false, e }
	}
	if Contents & which != 0 {
		oldf, e := ioutil.ReadFile(f.Name)
		if e != nil || string(oldf) != f.Contents { needswriting = true }
		if needswriting {
			myf, e := os.Open(f.Name+".new", os.O_WRONLY + os.O_TRUNC + os.O_CREAT + os.O_EXCL, f.Perms)
			if e != nil { return false, e }
			defer myf.Close()
			// To keep things simple, always set the ownership and permissions before writing...
			e = myf.Chown(f.Uid, f.Gid)
			if e != nil { return false, e }
			e = myf.Chmod(f.Perms)
			if e != nil { return false, e }
			
			n, e := io.WriteString(myf, f.Contents)
			if e != nil { return false, e }
			if n < len(f.Contents) { return false, io.ErrShortWrite }
			e = os.Rename(f.Name+".new", f.Name)
			if e != nil { return false, e }
		}
	}
	return needswriting, nil
}

func Stat(n string) (f Data, e os.Error) {
	f.Name = n
	stat, e := os.Lstat(f.Name)
	if e != nil { return }
	if !stat.IsRegular() { return f, os.NewError("Bad stuff....") }
	f.Perms = stat.Permission()
	f.Uid = stat.Uid
	f.Gid = stat.Gid
	return
}

func Read(n string) (f Data, e os.Error) {
	f, e = Stat(n)
	if e != nil { return }
	x, e := ioutil.ReadFile(f.Name)
	f.Contents = string(x)
	return
}

func (f Data) GoSet(which Field) {
	if which == 0 { return } // Nothing to set!
	ago.Import("github.com/droundy/goadmin/file")
	ago.Declare("var changed_files = make(map[string]bool)")
	ago.Declare("var file_changed = false")
	ago.Declare("var this_file_changed = false")
	code := "this_file_changed, e = file.Data{"+strconv.Quote(f.Name)+","
	if Contents & which != 0 {
		code += strconv.Quote(f.Contents) + ","
	} else {
		code += `"",`
	}
	code += fmt.Sprint(f.Uid, ",", f.Gid, ",", strconv.Itob(int(f.Perms), 8), "}.Set(", which, ")")
	code += fmt.Sprint("\n\tif e != nil { return }")
	code += fmt.Sprint("\n\tif this_file_changed { changed_files[`",f.Name,"`] = true }")
	code += fmt.Sprint("\n\tfile_changed = file_changed || this_file_changed")
	ago.Code(code)
}
