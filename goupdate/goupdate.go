package main

import (
	"fmt"
	"os"
	"path"
	"io"
	"io/ioutil"
  "strconv"
	"github.com/droundy/goadmin/ago/compile"
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
var privatekey crypt.PrivateKey
var publickey crypt.PublicKey
var sequence int64

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
	e := compile.Compile(*outname, fs)
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
	if e != nil { return }
	fi, err := os.Stat(*outname)
	if err != nil { return }
	enc0,err := os.Open(*outname+".encrypted", os.O_WRONLY + os.O_TRUNC + os.O_CREAT, 0644)
	if err != nil { return }
	enc, err := crypt.Encrypt(key, privatekey, enc0, fi.Size, sequence)
	if err != nil { return }
	plain,err := os.Open(*outname, os.O_RDONLY, 0644)
	if err != nil { return }
	_, err = io.Copyn(enc, plain, fi.Size)
	if err != nil { return }
}

func makeSource(name string) (err os.Error) {
	os.Remove(name) // just in case it already exists and has wrong
									// permissions (or even is already open by another
									// process)!
	out,err := os.Open(name, os.O_WRONLY + os.O_TRUNC + os.O_CREAT + os.O_EXCL, 0600) // will contain the key!
	if err != nil { return }
	defer out.Close()

	serialfile := *outname+".serial"
	privatefile := *outname+".private"
	publicfile := *outname+".public"
	if *keyfile != "FILENAME" {
		x,e := ioutil.ReadFile(*keyfile)
		if e != nil { return e }
		key = string(x)
		if len(*keyfile) > 4 && (*keyfile)[len(*keyfile)-4:] == ".key" {
			serialfile = (*keyfile)[0:len(*keyfile)-4] + ".serial"
			privatefile = (*keyfile)[0:len(*keyfile)-4] + ".private"
			publicfile = (*keyfile)[0:len(*keyfile)-4] + ".public"
		} else {
			serialfile = *keyfile + ".serial"
			privatefile = *keyfile + ".private"
			publicfile = *keyfile + ".public"
		}
	} else {
		x,e := ioutil.ReadFile(*outname+".key")
		if e != nil {
			key, err = crypt.CreateNewKey()
			if err != nil { return }
			ioutil.WriteFile(*outname+".key", []byte(key), 0600)
		} else {
			key = string(x)
		}
	}
	// Read or generate a serial number!
	x,e := ioutil.ReadFile(serialfile)
	if e == nil {
		// I don't particularly care if it doesn't exist, or if I have
		// trouble reading it...
		sequence, e = strconv.Atoi64(string(x))
		if e != nil {
			fmt.Println("Couldn't parse file", serialfile, "so starting at serial number 1")
			fmt.Println("File looks like: ", strconv.Quote(string(x)))
		} else {
			fmt.Println("Storing new serial number", sequence+1, "in file", serialfile)
		}
	} else {
		fmt.Println("No file", serialfile, "so starting at serial number 1")
	}
	sequence++;
	e = ioutil.WriteFile(serialfile, []byte(fmt.Sprint(sequence)), 0600)
	if e != nil { return e }
	// Read or generate a private key!
	x,e = ioutil.ReadFile(privatefile)
	y, e2 := ioutil.ReadFile(publicfile)
	if e == nil && e2 == nil {
		privatekey = crypt.PrivateKey(x)
		publickey = crypt.PublicKey(y)
	} else {
		fmt.Println("Creating a fresh RSA key pair, this may take some time...")
		publickey, privatekey, e = crypt.CreateRSAKeyPair()
		fmt.Println("Writing RSA key pair to disk...")
		if e != nil { return e }
		e = ioutil.WriteFile(privatefile, []byte(string(privatekey)), 0600)
		if e != nil { return e }
		e = ioutil.WriteFile(publicfile, []byte(string(publickey)), 0600)
		if e != nil { return e }
	}
	// Finally, let's write out the source code using this stuff.
	_, err = io.WriteString(out, `package main

import (
    "fmt"
    "os"
    "exec"
    "io"
    "http"
    "github.com/droundy/goopt"
    "github.com/droundy/goadmin/crypt"
    "github.com/droundy/goadmin/deps"
)

func main() {
    source := `)
	if err != nil { return }
	_, err = io.WriteString(out, strconv.Quote(path.Join(*source, *outname)))
	if err != nil { return }
	_, err = io.WriteString(out, `
    key := `)
	if err != nil { return }
	_, err = io.WriteString(out, strconv.Quote(key))
	if err != nil { return }
	_, err = io.WriteString(out, `
    var publickey crypt.PublicKey = `)
	if err != nil { return }
	_, err = io.WriteString(out, strconv.Quote(string(publickey)))
	if err != nil { return }
	_, err = io.WriteString(out, `
    var serialnum int64 = `)
	if err != nil { return }
	_, err = io.WriteString(out, strconv.Itoa64(sequence))
	if err != nil { return }
	_, err = io.WriteString(out, `
    exitwith := func (x string, e os.Error) {
        if e != nil {
            fmt.Fprintln(os.Stderr, x, e)
            os.Exit(1)
        }
    }
    exiton := func (e os.Error) {
        exitwith("Error updating: ", e)
    }
    update := func () (err os.Error) {
        fmt.Fprintln(os.Stderr, "I am trying to update from", source+".encrypted")
        outname,err := exec.LookPath(os.Args[0])
        exitwith("Couldn't find my own binary: ", err)
        var enc0 io.Reader
        isurl := false
        for _,c := range source {
            // This is a hokey way to check for URLs vs. local files.
            isurl = isurl || (c == ':')
        }
        if isurl {
            r, _, err := http.Get(source+".encrypted")
            exitwith("Couldn't retrieve encrypted source via http: ", err)
            enc0 = r.Body
            defer r.Body.Close()
        } else {
            enc0,err = os.Open(source+".encrypted", os.O_RDONLY, 0700)
            exitwith("Couldn't open encrypted file: ", err)
        }
        //fmt.Println("I have opened for reading", source+".encrypted")
        enc, mylen, newserialnum, err := crypt.Decrypt(key, publickey, enc0)
        exitwith("Couldn't decrypt file: ", err)
        if newserialnum <= serialnum {
          fmt.Println("New executable is same as the current one.")
        } else if newserialnum < serialnum {
          fmt.Println("New executable is older than the current one.")
        } else {
          plain,err := os.Open(outname+".new", os.O_WRONLY + os.O_TRUNC + os.O_CREAT, 0700)
          exiton(err)
	        defer plain.Close()
          //fmt.Println("I have opened for writing", outname)
          _, err = io.Copyn(plain, enc, mylen)
          exiton(err)
          plain.Close()
          exiton(os.Rename(outname+".new", outname))
          //fmt.Println("I have renamed", outname)
          //fmt.Println("I am updating...")
          exiton(os.Exec(outname, []string{os.Args[0]}, nil))
        }
        return
    }
    goopt.NoArg([]string{"--update"}, "update this executable", update)

    // The above is all effectively "init" stuff.
    goopt.Parse(func() []string { return []string{} })
    deps.Done()
}
`)
	return
}
