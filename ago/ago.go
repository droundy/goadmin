package ago

import (
	"os"
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
