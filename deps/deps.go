package deps

import (
	"os"
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

type Job <-chan Error

func WaitFor(js ...Job) os.Error {
	for _,j := range js {
		e := <- j
		if e != nil {
			return e // Quit early with error.
		}
	}
	return nil
}

func Run(f func () Error, js ...Job) Job {
	ch := make(chan Error)
	go func () {
		for _,j := range js {
			e := <- j
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
	return Job(ch)
}

func Execs(cmd string, args []string) os.Error {
	abscmd,err := exec.LookPath(cmd)
	if err != nil { return os.NewError("Couldn't find "+cmd+": "+err.String()) }
	
	cmdargs := make([]string, len(args)+1)
	cmdargs[0] = cmd
	for i,a := range args {
		cmdargs[i+1] = a
	}
	os.Stdout.Write([]byte(abscmd+" ...\n"))
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
