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
	motd.GoSet(file.All)
	file.File{file.StatData{"/tmp/test-from-asroot", 0, 0, 0666, file.IsFile}, "Some demo contents\n"}.GoSet(file.All)
	ago.Code(`if file_changed { fmt.Println("I changed a file!") }`)
	ago.Print("files")
}
