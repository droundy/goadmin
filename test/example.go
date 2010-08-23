package main

import (
	"fmt"
	"github.com/droundy/goadmin/deps"
)

var hello = deps.Run(func () deps.Error { fmt.Println("Hello world!"); return nil })
