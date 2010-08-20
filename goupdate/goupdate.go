package main

import (
	"fmt"
	"os"
	"path"
	"io"
	"io/ioutil"
	"github.com/droundy/goadmin/ago"
	"github.com/droundy/goadmin/crypt"
	"github.com/droundy/goopt"
)

var urlbase = goopt.String([]string{"--url"}, "", "the base of the URL to download from")

var outname = goopt.String([]string{"--output"}, "FILENAME",	"the name of the output file")

var source = goopt.String([]string{"--source"}, func () string {
	wd, _ := os.Getwd()
	return wd
}(),	"the url where updates will be available")

var keyfile = goopt.String([]string{"--keyfile"}, "FILENAME", "the name of a key file")
var key = ""

func main() {
	goopt.Parse(func() []string { return []string{} })

	if len(goopt.Args) < 1 {
		fmt.Println("You need to provide a go file argument.")
		os.Exit(1);
	}
	if *outname == "FILENAME" {
		*outname = goopt.Args[0][0:len(goopt.Args[0])-3]
	}
	fs := make([]string, len(goopt.Args)+1)
	for i,f := range goopt.Args {
		fs[i] = f
	}
	fs[len(goopt.Args)] = "testing.go"
	makeSource("testing.go")
	ioutil.WriteFile(*outname+".key", []byte(key), 0600)
	e := ago.Compile(*outname, fs)
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
	e = os.Chmod(*outname, 0700) // So noone bad can read the embedded key.
	if e != nil { return }
	fi, err := os.Stat(*outname)
	if err != nil { return }
	enc0,err := os.Open(*outname+".encrypted", os.O_WRONLY + os.O_TRUNC + os.O_CREAT, 0644)
	if err != nil { return }
	enc, err := crypt.Encrypt(key, enc0, fi.Size)
	if err != nil { return }
	plain,err := os.Open(*outname, os.O_RDONLY, 0644)
	if err != nil { return }
	_, err = io.Copyn(enc, plain, fi.Size)
	if err != nil { return }
}

func makeSource(name string) (err os.Error) {
	out,err := os.Open(name, os.O_WRONLY + os.O_TRUNC + os.O_CREAT, 0644)
	if err != nil { return }
	defer out.Close()

	if *keyfile != "FILENAME" {
		x,e := ioutil.ReadFile(*keyfile)
		if e != nil { return e }
		key = string(x)
	} else {
		key, err = crypt.CreateNewKey()
		if err != nil { return }
	}
	_, err = io.WriteString(out, `package main

import (
    "fmt"
    "os"
    "exec"
    "io"
    "http"
    "github.com/droundy/goopt"
    "github.com/droundy/goadmin/crypt"
)

func init() {
    fmt.Println("This is only a test.\n")

    source := "`)
	if err != nil { return }
	_, err = io.WriteString(out, path.Join(*source, *outname))
	if err != nil { return }
	_, err = io.WriteString(out, `"

    key := string([]byte{
`)
	if err != nil { return }
	for i:=0; i<len(key); i++ {
		_, err = fmt.Fprintln(out, "    ", key[i], ",")
		if err != nil { return }
	}
	_, err = io.WriteString(out, `
})

    decrypt := func (f string) (err os.Error) {
        outname := "plaintext"
        if f[len(f)-len(".encrypted"):] == ".encrypted" {
            outname = f[0:len(f)-len(".encrypted")]+".new"
        }
        enc0,err := os.Open(f, os.O_RDONLY, 0644)
        if err != nil { return }
        plain,err := os.Open(outname, os.O_WRONLY + os.O_TRUNC + os.O_CREAT, 0644)
        if err != nil { return }
	      defer plain.Close()
        enc, mylen, err := crypt.Decrypt(key, enc0)
        if err != nil { return }
        _, err = io.Copyn(plain, enc, mylen)
        if err != nil { return }
        fmt.Println("I have decrypted it.")
        os.Exit(0)
        return
    }
    exiton := func (e os.Error) {
        if e != nil {
            fmt.Fprintln(os.Stderr, "Error updating: ", e)
            os.Exit(1)
        }
    }
    update := func () (err os.Error) {
        fmt.Println("I am trying to update from", source+".encrypted")
        outname,err := exec.LookPath(os.Args[0])
        exiton(err)
        var enc0 io.Reader
        isurl := false
        for _,c := range source {
            // This is a hokey way to check for URLs vs. local files.
            isurl = isurl || (c == ':')
        }
        if isurl {
            r, _, err := http.Get(source+".encrypted")
            exiton(err)
            enc0 = r.Body
            defer r.Body.Close()
        } else {
            enc0,err = os.Open(source+".encrypted", os.O_RDONLY, 0700)
            exiton(err)
        }
        //fmt.Println("I have opened for reading", source+".encrypted")
        err = os.Rename(outname, outname+".old")
        //fmt.Println("I have renamed", outname)
        plain,err := os.Open(outname, os.O_WRONLY + os.O_TRUNC + os.O_CREAT, 0700)
        exiton(err)
	      defer plain.Close()
        //fmt.Println("I have opened for writing", outname)
        enc, mylen, err := crypt.Decrypt(key, enc0)
        exiton(err)
        _, err = io.Copyn(plain, enc, mylen)
        exiton(err)
        plain.Close()
        exiton(os.Remove(outname+".old"))
        //fmt.Println("I am updating...")
        err = os.Exec(outname, []string{os.Args[0]}, nil)
        exiton(err)
        return
    }
    goopt.ReqArg([]string{"--decrypt"}, "FILENAME", "decrypt a file", decrypt)
    goopt.NoArg([]string{"--update"}, "update this executable", update)
}
`)
	return
}
