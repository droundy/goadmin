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

Why?
----

Puppet is nice, but it is simultaneously too powerful and too weak for
my uses.  It's got all sorts of cross-platform features that I have no
use for, and as a result, it doesn't work particularly well on Debian
(and Debian-derived) systems.  Also, puppet is an *amazing* resource
hog.
