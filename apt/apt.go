package apt

import (
	"os"
	"github.com/droundy/goadmin/deps"
)

func Install(pkg string) os.Error {
	return runapt(func() os.Error {
		return deps.Exec("apt-get", "install", "-y", pkg)
	})
}

func Remove(pkg string) os.Error {
	return runapt(func() os.Error {
		return deps.Exec("apt-get", "remove", "-y", pkg)
	})
}

func Update() os.Error {
	return runapt(func() os.Error {
		return deps.Exec("apt-get", "update", "-y")
	})
}

func Upgrade() os.Error {
	return runapt(func() os.Error {
		return deps.Exec("apt-get", "upgrade", "-y")
	})
}

func AutoClean() os.Error {
	return runapt(func() os.Error {
		return deps.Exec("apt-get", "autoclean")
	})
}

func DistUpgrade() os.Error {
	return runapt(func() os.Error {
		return deps.Exec("apt-get", "dist-upgrade", "-y")
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
