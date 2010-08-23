- Add support for signing encrypted files with RSA, and embedding the
  public key in the executable.  This will mean that grabbing the
  executable won't allow someone to pretend to be the server.

- Add support for an enumeration of executables, to get around replay
  attacks (i.e. where someone presents an old, buggy version of an
  admin executable).
