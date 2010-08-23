# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

include $(GOROOT)/src/Make.$(GOARCH)

asroot: pw.go upgrade.go
	../../goupdate/goupdate --output asroot $^

%.go: %.go.go
	../../secretrun/secretrun $< > $@

clean:
	rm -f *.[568] asroot asroot.encrypted pw.go # Not the key
