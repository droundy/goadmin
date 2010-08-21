package apt

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"github.com/droundy/goadmin/deps"
)

func Install(pkg string) os.Error {
	return runapt(func() os.Error {
		once.Do(findInstalled)
		if _,i := installed[pkg]; !i {
			return deps.Exec("apt-get", "install", "-y", "-q", pkg)
		}
		fmt.Println(pkg, "is already installed.")
		return nil
	})
}

func Remove(pkg string) os.Error {
	return runapt(func() os.Error {
		once.Do(findInstalled)
		if _,i := installed[pkg]; i {
			return deps.Exec("apt-get", "remove", "-y", "-q", pkg)
		}
		fmt.Println(pkg, "is already removed.")
		return nil
	})
}

func Update() os.Error {
	return runapt(func() os.Error {
		return deps.Exec("apt-get", "update", "-qq")
	})
}

func Upgrade() os.Error {
	return runapt(func() os.Error {
		return deps.Exec("apt-get", "upgrade", "-y", "-q")
	})
}

func AutoClean() os.Error {
	return runapt(func() os.Error {
		return deps.Exec("apt-get", "autoclean", "-qq")
	})
}

func DistUpgrade() os.Error {
	return runapt(func() os.Error {
		return deps.Exec("apt-get", "dist-upgrade", "-y", "-q")
	})
}

func runapt(f func () os.Error) os.Error {
	x := make(chan os.Error)
	apts <- func () {
		x <- f()
	}
	return <- x
}
var apts = make(chan func ())

var installed = make(map[string]struct{})

var once sync.Once
func findInstalled() {
	// dpkg-query -W -f='${Package} ${Status}\n'
	st, err := deps.ExecRead("dpkg-query", "-W", "-f=${Package} ${Status}\\n")
	if err == nil {
		matches := regexp.MustCompile("([^\n]+) install ok installed\n").FindAllSubmatch(st, -1)
		for _,match := range matches {
			//fmt.Println("Found package", string(match[1]), len(match[1]))
			installed[string(match[1])] = struct{}{}
		}
	}
	// dpkg-query -W -f='${Provides} ${Status}\n'
	st, err = deps.ExecRead("dpkg-query", "-W", "-f=${Provides} ${Status}\\n")
	if err == nil {
		matches := regexp.MustCompile("([^\n]+) install ok installed\n").FindAllSubmatch(st, -1)
		for _,match := range matches {
			submatches := regexp.MustCompile("[^, ]+").FindAll(match[1], -1)
			for _, submatch := range submatches {
				//fmt.Println("Found provides:", string(submatch))
				installed[string(submatch)] = struct{}{}
			}
		}
	}
}

func init() {
	// If we include this package, we probably want to do an update.
	//apts <- func () { deps.Exec("apt-get", "update", "-y") }
	go func () {
		for {
			f := <- apts
			f()
		}
	}()
}
