- Add support for signing encrypted files with RSA, and embedding the
  public key in the executable.  This will mean that grabbing the
  executable won't allow someone to pretend to be the server.

- Add support for an enumeration of executables, to get around replay
  attacks (i.e. where someone presents an old, buggy version of an
  admin executable).

- Make the dependency framework actually useful, and decide if it
  gains us anything.  The main purpose is to allow modules to
  interact, via variable initializations, so code can be shared
  between different targets.

- Add a way to ensure that files are present.

- Add a code-generator?
