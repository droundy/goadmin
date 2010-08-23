package deps

import (
	"os"
	"fmt"
	"io/ioutil"
	"exec"
	"strconv"
)

func NewError(e string) Error {
	return os.NewError(e)
}

type Error interface {
	String() string
}

func ExitWith(m string, e os.Error) {
  if e != nil {
		os.Stderr.Write([]byte(m+": "+e.String()))
    os.Exit(1)
  }
}

var alljobs = make([]Job, 0, 12)
func init() {
	go func() {
		// collect all the jobs and wait for them to finish
		for {
			x := <- reportjob
			if len(alljobs) == cap(alljobs) {
				n := make([]Job, len(alljobs), 2*cap(alljobs))
				for i,j := range alljobs {
					n[i] = j
				}
				alljobs = n
			}
			alljobs = alljobs[0:len(alljobs)+1]
			alljobs[len(alljobs)-1] = x
		}
	}()
}

//type Job <-chan Error
type Job struct {
	Ch <-chan Error
}
var reportjob = make(chan Job)

func waitFors(js []Job) os.Error {
	for _,j := range js {
		e := <- j.Ch
		if e != nil {
			return e // Quit early with error.
		}
	}
	return nil
}

func WaitFor(js ...Job) os.Error {
	return waitFors(js)
}

func Done() os.Error {
	return waitFors(alljobs)
}

func Run(f func () Error, js ...Job) Job {
	ch := make(chan Error)
	go func () {
		for _,j := range js {
			e := <- j.Ch
			if e != nil {
				for {
					ch <- e // This is a very hokey way of broadcasting...
				}
			}
		}
		e := f()
		for {
			ch <- e // This is a very hokey way of broadcasting...
		}
	}()
	j := Job{ch}
	go func() { reportjob <- j }()
	return j
}

func ExecRead(cmd string, args ...string) (out []byte, err os.Error) {
	abscmd,err := exec.LookPath(cmd)
	if err != nil { return out, os.NewError("Couldn't find "+cmd+": "+err.String()) }
	
	cmdargs := make([]string, len(args)+1)
	cmdargs[0] = cmd
	for i,a := range args {
		cmdargs[i+1] = a
	}
	printexec(cmd, args)

	pid,err := exec.Run(abscmd, cmdargs, nil, "",
		exec.PassThrough, exec.Pipe, exec.PassThrough)
	if err != nil { return }
	out,err = ioutil.ReadAll(pid.Stdout)
	if err != nil { return }
	ws,err := pid.Wait(0) // could have been os.WRUSAGE
	if err != nil { return }
	if ws.ExitStatus() != 0 {
		err = os.NewError(cmd+" exited with status "+strconv.Itoa(ws.ExitStatus()))
	}
	return out, nil
}

func printexec(cmd string, args []string) {
	fmt.Print(cmd)
	for _,a := range args {
		fmt.Print(" ", a)
	}
	fmt.Println()
}

func Execs(cmd string, args []string) os.Error {
	abscmd,err := exec.LookPath(cmd)
	if err != nil { return os.NewError("Couldn't find "+cmd+": "+err.String()) }
	
	cmdargs := make([]string, len(args)+1)
	cmdargs[0] = cmd
	for i,a := range args {
		cmdargs[i+1] = a
	}

	printexec(cmd, args)
	pid, err := exec.Run(abscmd, cmdargs, nil, "",
		exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil { return err }
	wmsg,err := pid.Wait(0)
	if err != nil { return err }
	if wmsg.ExitStatus() != 0 {
		return os.NewError(cmd+" exited with status "+strconv.Itoa(wmsg.ExitStatus()))
	}
	return nil
}

func Exec(cmd string, args ...string) os.Error {
	return Execs(cmd, args)
}
