package main

import (
	"github.com/droundy/goadmin/apt"
	"github.com/droundy/goadmin/deps"
)

var upgrades = deps.Run(func () (e deps.Error) {
	e = apt.Update()
	if e != nil { return }
	e = apt.Upgrade()
	if e != nil { return }
	e = apt.Install("chromium-browser")
	if e != nil { return }
	e = apt.Install("finger")
	if e != nil { return }
	e = apt.Remove("xmonad")
	if e != nil { return }
	apt.AutoClean() // I don't care if this fails
	return
})
