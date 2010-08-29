package main

import (
	"github.com/droundy/goadmin/hosts"
	"github.com/droundy/goadmin/ago"
)

func main() {
	// Want all machines to have myself and guest (and have the
	// passwords copied over, if this is run as root, so the passwords
	// are available from shadow).
	hosts.Get()["bingley"].GoSet()
	ago.Print("hostfile")
}
