# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

all:
	goinstall github.com/droundy/goopt
	cd crypt; make install
	cd deps; make install
	cd apt; make install
	cd ago; make install
	cd goupdate; make

clean:
	cd crypt; make clean
	cd deps; make clean
	cd apt; make clean
	cd ago; make clean
	cd goupdate; make clean
