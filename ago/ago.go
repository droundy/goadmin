package ago

import (
	"fmt"
	"os"
	"exec"
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
	e = justrun(archnum()+"g", args)
	if e != nil { return }
	return justrun(archnum()+"l", []string{"-o", outname, objname})
}

func justrun(cmd string, args []string) os.Error {
	abscmd,err := exec.LookPath(cmd)
	if err != nil { return os.NewError("Couldn't find "+cmd+": "+err.String()) }
	
	cmdargs := make([]string, len(args)+1)
	cmdargs[0] = cmd
	for i,a := range args {
		cmdargs[i+1] = a
	}
	pid, err := exec.Run(abscmd, cmdargs, nil, "",
		exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil { return err }
	wmsg,err := pid.Wait(0)
	if err != nil { return err }
	if wmsg.ExitStatus() != 0 {
		return os.NewError(cmd+" exited with status "+fmt.Sprint(wmsg.ExitStatus()))
	}
	return nil
}
