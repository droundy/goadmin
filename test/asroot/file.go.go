package main

import (
	"github.com/droundy/goadmin/file"
	"github.com/droundy/goadmin/ago"
)

func main() {
	// Want all machines to have the same /etc/motd
	motd,_ := file.Read("/etc/motd")
	motd.MakeCodeToSet(file.All)
	file.Data{"/tmp/test-from-asroot", "Some demo contents\n", 0, 0, 0666}.MakeCodeToSet(file.All)
	ago.Print("files")
}
