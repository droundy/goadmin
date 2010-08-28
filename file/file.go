package file

import (
	"os"
	"io"
	"path"
	"io/ioutil"
	"fmt"
	"strconv"
	"github.com/droundy/goadmin/ago"
	"github.com/droundy/goadmin/gomakefile"
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

type StatMode int
const (
  IsFile StatMode = iota
	IsDirectory
	IsSymlink
	IsOther
)

type StatData struct {
	Name string
	Uid, Gid int
	Perms uint32
	Mode StatMode
}

type File struct {
	StatData
	Contents string
}

type Directory struct {
	StatData
	Contents map[string]StatEntry
}

type StatEntry interface {
	Set(Field) (needswriting bool, e os.Error)
	GoSet(Field)
	Move(from, to string)
}

func (f *StatData) Move(from, to string) {
	from = path.Clean(from)+"/"
	to = path.Clean(to)
	if f.Name == "" {
		// This is silly!
		return
	}
	if from == "./" {
		if f.Name[0] != '/' {
			f.Name = path.Join(to, f.Name)
		}
	} else if len(f.Name) > len(from) && f.Name[0:len(from)] == from {
		f.Name = path.Join(to, f.Name[len(from):])
	}
}

func (f *Directory) Move(from, to string) {
	f.StatData.Move(from, to)
	for x := range f.Contents {
		f.Contents[x].Move(from, to)
	}
}

// *StatData Set modifies the StatData to reflect that actual stats of
// the file, when it isn't instructed to set those values.  Thus it is
// an input/output method that can be used to ensure that all the
// stats of the *StatData are correct, based on the current stats of
// the file or directory.

func (f *StatData) Set(which Field) (needswriting bool, e os.Error) {
	stat, e := Stat(f.Name)
	if e != nil {
		_, ispatherror := e.(*os.PathError) // *os.PathError means file doesn't exist (which is okay)
		if e != nil && !ispatherror { return false, e }
		return true, nil
	}
	if stat.Mode != f.Mode {
		return true, os.Remove(f.Name)
	}
	if Perms & which != 0 && stat.Perms != f.Perms {
		e = os.Chmod(f.Name, f.Perms)
		if e != nil { return false, e }
	} else {
		f.Perms = stat.Perms
	}
	// Set the ownership if needed...
	if Gid & which == 0 { f.Gid = stat.Gid }
	if Uid & which == 0 { f.Uid = stat.Uid }
	if Uid & which != 0 || Gid & which != 0 {
		e = os.Lchown(f.Name, f.Uid, f.Gid)
		if e != nil { return false, e }
	}
	return false, nil
}

func (ent *StatData) Read() (f StatEntry, e os.Error) {
	switch ent.Mode {
	case IsFile:
		ff, e := ent.readFile()
		if e != nil {
			return nil, e
		} else {
			return &ff, e
		}
	case IsDirectory:
		ff, e := ent.readDirectory()
		if e != nil {
			return nil, e
		} else {
			return &ff, e
		}
	default:
		return nil, os.NewError("I only know how to read files and directories so far.")
	}
	return
}

func (st *StatData) readFile() (f File, e os.Error) {
	f.StatData = *st
	if f.Mode != IsFile {
		return f, os.NewError("Not a regular file:  "+f.Name)
	}
	x, e := ioutil.ReadFile(f.Name)
	f.Contents = string(x)
	return
}

func (st *StatData) readDirectory() (f Directory, e os.Error) {
	f.StatData = *st
	if f.Mode != IsDirectory {
		return f, os.NewError("Not a directory:  "+f.Name)
	}
	x, e := os.Open(f.Name, os.O_RDONLY, f.Perms)
	if e != nil { return }
	fs, e := x.Readdirnames(-1)
	f.Contents = make(map[string]StatEntry)
	for _, fname := range fs {
		if fname != "." && fname != ".." {
			ent, e := Stat(path.Join(f.Name, fname))
			if e != nil { return f, e }
			xxx, _ := ent.Read()
			f.Contents[fname] = xxx
		}
	}
	return
}

func (f *Directory) Set(which Field) (needswriting bool, e os.Error) {
	needswriting, e = f.StatData.Set(which)
	if e != nil { return }
	if needswriting {
		// This means it doesn't currently exist...
		e = os.Mkdir(f.Name, f.Perms)
		if e != nil { return }
		// Set owner...
		e = os.Lchown(f.Name, f.Uid, f.Gid)
		if e != nil { return false, e }
		// Chmod it in any case, because umask may have modified the
		// desired permissions.
		e = os.Chmod(f.Name, f.Perms)
		if e != nil { return false, e }
	}
	if Contents & which != 0 {
		for _, c := range f.Contents {
			// Set all the contents as well!
			c.Set(which)
		}
	}
	return
}

func (f *Directory) GoSet(which Field) {
	if which == 0 { return } // Nothing to set!
	ago.Import("github.com/droundy/goadmin/file")
	ago.Declare("var changed_files = make(map[string]bool)")
	ago.Declare("var file_changed = false")
	ago.Declare("var this_file_changed = false")
	ago.Declare("var myd file.Directory")
	code := "myd = file.Directory{file.StatData{"+strconv.Quote(f.Name)+","
	code += fmt.Sprint(f.Uid, ",", f.Gid, ", 0", strconv.Itob(int(f.Perms), 8), ", file.IsFile},")
	code += `map[string]file.StatEntry{}}`
	code += fmt.Sprint("\n\tthis_file_changed, e = myd.Set(", which, ")")
	code += fmt.Sprint("\n\tif e != nil { return }")
	code += fmt.Sprint("\n\tif this_file_changed { changed_files[`",f.Name,"`] = true }")
	code += fmt.Sprint("\n\tfile_changed = file_changed || this_file_changed")
	ago.Code(code)
	if Contents & which != 0 {
		for _,x := range f.Contents {
			x.GoSet(which)
		}
	}
}

func (f *File) Set(which Field) (needswriting bool, e os.Error) {
	needswriting, e = f.StatData.Set(which)
	if e != nil { return }
	{
		// Let's clear out any existing ".new" file
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

func Stat(n string) (f StatData, e os.Error) {
	f.Name = n
	stat, e := os.Lstat(f.Name)
	if e != nil { return }
	if stat.IsRegular() {
		f.Mode = IsFile
	} else if stat.IsDirectory() {
		f.Mode = IsDirectory
	} else if stat.IsSymlink() {
		f.Mode = IsSymlink
	} else {
		f.Mode = IsOther
	}
	f.Perms = stat.Permission()
	f.Uid = stat.Uid
	f.Gid = stat.Gid
	gomakefile.AddDep(f.Name)
	return
}

func Read(n string) (f StatEntry, e os.Error) {
	st, e := Stat(n)
	if e != nil { return }
	return st.Read()
}

func (f *File) GoSet(which Field) {
	if which == 0 { return } // Nothing to set!
	ago.Import("github.com/droundy/goadmin/file")
	ago.Declare("var changed_files = make(map[string]bool)")
	ago.Declare("var file_changed = false")
	ago.Declare("var this_file_changed = false")
	ago.Declare("var myf file.File")
	code := "myf = file.File{file.StatData{"+strconv.Quote(f.Name)+","
	code += fmt.Sprint(f.Uid, ",", f.Gid, ", 0", strconv.Itob(int(f.Perms), 8), ", file.IsFile},")
	if Contents & which != 0 {
		code += strconv.Quote(f.Contents) + "}"
	} else {
		code += `""}`
	}
	code += fmt.Sprint("\n\tthis_file_changed, e = myf.Set(", which, ")")
	code += fmt.Sprint("\n\tif e != nil { fmt.Println(`Trouble setting",f.Name,":`,e); return }")
	code += fmt.Sprint("\n\tif this_file_changed { changed_files[`",f.Name,"`] = true }")
	code += fmt.Sprint("\n\tfile_changed = file_changed || this_file_changed")
	ago.Code(code)
}
