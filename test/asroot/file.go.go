package main

import (
	"github.com/droundy/goadmin/file"
	"github.com/droundy/goadmin/ago"
)

func main() {
	ago.Import("fmt")
	ago.Code(`fmt.Println("I am thinking about changing some files...")`)
	// Want all machines to have the same /etc/motd
	motd,_ := file.Read("/etc/motd")
	motd.MakeCodeToSet(file.All)
	file.Data{"/tmp/test-from-asroot", "Some demo contents\n", 0, 0, 0666}.MakeCodeToSet(file.All)
	ago.Code(`if file_changed { fmt.Println("I changed a file!") }`)
	ago.Print("files")
}
