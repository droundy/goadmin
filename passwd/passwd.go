package passwd

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sync"
	"strconv"
)

type User struct {
	Name, Passwd string
	Uid, Gid int
	Comment, Home, Shell string
}

var passwd = make(map[string]User)
var once sync.Once
func Get() map[string]User {
	once.Do(func () {
		x, e := ioutil.ReadFile("/etc/passwd")
		if e == nil {
			pwent := "([^:\n]+):([^:]+):([0-9]+):([0-9]+):([^:]*):([^:]+):([^:]+)\n"
			matches := regexp.MustCompile(pwent).FindAllSubmatch(x, -1)
			for _,match := range matches {
				username := string(match[1])
				passwr := string(match[2])
				uid, e := strconv.Atoi(string(match[3]))
				if e != nil {
					fmt.Println("Error reading uid of", username, string(match[3]))
					continue
				}
				gid, e := strconv.Atoi(string(match[4]))
				if e != nil {
					fmt.Println("Error reading gid of", username, string(match[4]))
					continue
				}
				comment := string(match[5])
				home := string(match[6])
				shell := string(match[7])
				passwd[username] = User{username, passwr, uid, gid, comment, home, shell}
			}
			x, e = ioutil.ReadFile("/etc/shadow")
			if e == nil {
				pwent := "([^:\n]+):([^:]+):[^:]*:[^:]*:[^:]*:[^:]*:[^:]*:[^:]*:[^:]*\n"
				matches := regexp.MustCompile(pwent).FindAllSubmatch(x, -1)
				for _,match := range matches {
					username := string(match[1])
					passwr := string(match[2])
					if us, ok := passwd[username]; ok {
						us.Passwd = passwr // update the password!
						passwd[username] = us
					}
				}
			} else {
				fmt.Println("Unable to read /etc/shadow.")
			}
		}
	})
	return passwd
}
