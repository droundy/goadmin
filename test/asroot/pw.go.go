package main

import (
	"github.com/droundy/goopt"
	"github.com/droundy/goadmin/passwd"
	"github.com/droundy/goadmin/ago"
)

func main() {
	goopt.Parse(func() []string { return []string{} })

	// Want all machines to have myself and guest (and have the
	// passwords copied over, if this is run as root, so the passwords
	// are available from shadow).
	passwd.Get()["droundy"].MakeCodeToSet(passwd.All)
	passwd.Get()["guest"].MakeCodeToSet(passwd.All)
	ago.Print("droundy")
}
