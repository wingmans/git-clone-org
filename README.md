# git-clone-org

Clones all repositories from a github org or a github user into the current folder

````
Usage: git-clone-org <mode> <filter> [flags]

Clone all repositories from a GH org or user

Arguments:
  <mode>      Mode of operation (usr, org).
  <filter>    Filter criteria.

Flags:
  -h, --help               Show context-sensitive help.
  -c, --clean              Perform a clean operation.  TBD
  -n, --noop               No operation mode.
  -v, --verbose=COUNTER    Increase verbosity by repeating. -v, -vv, -vvv.
  ```

  ## Examples

  ```
  git-clone-org usr wingmans (Clones all repos from the user wingmans)