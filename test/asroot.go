package main

import (
	"fmt"
	"github.com/droundy/goopt"
	"github.com/droundy/goadmin/apt"
	"github.com/droundy/goadmin/deps"
	"github.com/droundy/goadmin/passwd"
)

func main() {
	goopt.Parse(func() []string { return []string{} })

	fmt.Println("This is David's demo admin script.")
	fmt.Println("\nIt will gradually grow to showcase new features.")
	
	deps.ExitWith("Error updating", apt.Update())
	deps.ExitWith("Error upgrading", apt.Upgrade())
	deps.ExitWith("Error installing", apt.Install("chromium-browser"))
	deps.ExitWith("Error installing", apt.Install("finger"))
	deps.ExitWith("Error removing", apt.Remove("xmonad"))

	u := passwd.Get()["droundy"]
	u.Gid = 100
	u.Set(passwd.All)
	apt.AutoClean() // I don't care if this fails!
}
