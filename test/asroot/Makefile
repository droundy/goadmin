# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

include $(GOROOT)/src/Make.$(GOARCH)

asroot: upgrade.go pw.generated.go file.generated.go hosts.generated.go
	../../goupdate/goupdate --output asroot $^

%.generated.go: %.go.go
	../../secretrun/secretrun -o $@ $<

clean:
	rm -f *.[568] asroot asroot.encrypted *.generated.go # Not the key

file.generated.go: /etc/motd
pw.generated.go: /etc/passwd
hosts.generated.go: /etc/hosts
