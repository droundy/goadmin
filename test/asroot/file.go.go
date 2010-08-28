package main

import (
	"fmt"
	"github.com/droundy/goadmin/file"
	"github.com/droundy/goadmin/ago"
)

func main() {
	ago.Import("fmt")
	ago.Code(`fmt.Println("I am thinking about changing some files...")`)
	// Want all machines to have the same /etc/motd
	motd,_ := file.Read("/etc/motd")
	motd.GoSet(file.All)
	tmp := file.File{file.StatData{"/tmp/test-from-asroot", 0, 0, 0666, file.IsFile}, "Some demo contents\n"}
	tmp.GoSet(file.All)

	demofiles,e := file.Read("demo-files")
	if e == nil {
		fmt.Println("I am going output demo-files");
		demofiles.Move(".", "/tmp")
		demofiles.GoSet(file.All)
	} else {
		fmt.Println("Error reading demo-files: ", e)
	}

	subdir,e := file.Read("demo-files/subdir")
	if e == nil {
		fmt.Println("I am going output demo-files/subdir");
		subdir.Move("demo-files", "/tmp")
		subdir.GoSet(file.All)
	} else {
		fmt.Println("Error reading subdir: ", e)
	}

	ago.Code(`if file_changed { fmt.Println("I changed a file!") }`)
	ago.Print("files")
}
