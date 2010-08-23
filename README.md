goadmin
=======

Goadmin is a suite of tools intended to (eventually, if I don't get
tired of it first) serve as an alternative to
[puppet](http://www.puppetlabs.com), but without the cross-platform
capabilities.

What?
-----

The basic idea is that you write a program in go that you would like
to have periodically run on your system.  Goadmin will provide
libraries to help you write this program.

The tricky part is how to easily distribute this program to all the
machines that you administer.  The approach that goadmin takes is to
modify your admin program to embed a key into it, along with code to
fetch and decrypt (using said key) an updated binary.  So you just
need to put `/path/to/myadmin --update`, and your program will be run
periodically, and updated when a change is made.

How?
----

Your admin program consists of one or more files that define variables
that are initialized using `deps.Run`, which determine what needs to
be done.  You compile it with

    goupdate --source=URL_WHERE_YOU_PUT_ENCRYPTED_BINARIES your-admin-code.go otherfile.go

This will compile your code, and generate a key (called
`your-admin-code.key`), as well as a plain-text and encrypted binary.
Use `scp` to put the binary on the computer (or computers) to be
adminned.  Then when you make a change, you just need to recompile
(using the same command line as above) to generate a new binary
encrypted with the same key.  Put it at the URL specified, and when
you run

    ./your-admin-code --update

the program will grab the new version, replace itself, and exec it.

Note that since only symmetric encryption is used, if someone bad gets
a copy of the unencrypted `your-admin-code`, that someone has access
to the key and can impersonate the server to that client or any other
that shares the same key.  I plan on adding signing with a public key.

Goadmin also tracks serial numbers of executables, to get around
replay attacks (i.e. where someone presents an old, possibly buggy
version of an admin executable).

Why?
----

Puppet is nice, but it is simultaneously too powerful and too weak for
my uses.  It's got all sorts of cross-platform features that I have no
use for, and as a result, it doesn't work particularly well on Debian
(and Debian-derived) systems.  Also, puppet is an *amazing* resource
hog.
