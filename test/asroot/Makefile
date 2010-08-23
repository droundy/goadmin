# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

include $(GOROOT)/src/Make.$(GOARCH)

asroot: pw.go upgrade.go
	../../goupdate/goupdate --output asroot $^

%.go: %.go.go
	$(GC) -o _go_.$(O) $<
	$(LD) -o _go_ _go_.$(O)
	./_go_ > $@

clean:
	rm -f *.[568] _go_* asroot asroot.encrypted pw.go # Not the key
