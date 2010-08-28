# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

all:
	goinstall github.com/droundy/goopt
	cd crypt; make install
	cd deps; make install
	cd ago/compile; make install
	cd ago; make install
	cd passwd; make install
	cd file; make install
	cd apt; make install
	cd ago; make install
	cd goupdate; make
	cd secretrun; make

clean:
	cd crypt; make clean
	cd deps; make clean
	cd ago/compile; make install
	cd ago; make clean
	cd passwd; make clean
	cd file; make clean
	cd apt; make clean
	cd goupdate; make clean
	cd secretrun; make clean
	cd test/asroot; make clean
