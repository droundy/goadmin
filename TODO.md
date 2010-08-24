- Add handler for /etc/hosts

- Don't bother rewriting files that haven't changed.

- Add mechanism for noticing if files have changed.

- Add mechanism for updating Makefile dependencies (e.g. passwd could
  add a dependency on /etc/passwd and /etc/shadow, or file could add
  dependencies on the relevant files).
