package main

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"encoding/binary"
	"github.com/droundy/goadmin/ago"
	"github.com/droundy/goadmin/crypt"
	"github.com/droundy/goopt"
)

var urlbase = goopt.String([]string{"--url"},
	"", "the base of the URL to download from")

var outname = goopt.String([]string{"--output"},
	"FILENAME",	"the name of the output file")

var keyfile = goopt.String([]string{"--keyfile"},
	"FILENAME",	"the name of a key file")
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
	fi, err := os.Stat(*outname)
	if err != nil { return }
	enc0,err := os.Open(*outname+".encrypted", os.O_WRONLY + os.O_TRUNC + os.O_CREAT, 0644)
	if err != nil { return }
	enc, err := crypt.Encrypt(key, enc0)
	if err != nil { return }
	err = binary.Write(enc, binary.LittleEndian, fi.Size) // First store the file size
	if err != nil { return }
	plain,err := os.Open(*outname, os.O_RDONLY, 0644)
	if err != nil { return }
	_, err = io.Copyn(enc, plain, fi.Size)
	if err != nil { return }
	io.WriteString(enc,"this is just some padding, only a few dozen bytes.\n");
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
    "io"
    "encoding/binary"
    "github.com/droundy/goopt"
    "github.com/droundy/goadmin/crypt"
)

var ourkey = string([]byte{
`)
	if err != nil { return }
	for i:=0; i<len(key); i++ {
		_, err = fmt.Fprintln(out, "    ", key[i], ",")
		if err != nil { return }
	}
	_, err = io.WriteString(out, `
})

func init() {
    fmt.Println("This is only a test.\n")
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
        enc, err := crypt.Decrypt(ourkey, enc0)
        if err != nil { return }
        var mylen int64
        err = binary.Read(enc, binary.LittleEndian, &mylen) // First read the file size
        if err != nil { return }
        _, err = io.Copyn(plain, enc, mylen)
        if err != nil { return }
        os.Exit(0)
        return
    }
    goopt.ReqArg([]string{"--decrypt"}, "FILENAME", "decrypt a file", decrypt)
}
`)
	return
}
