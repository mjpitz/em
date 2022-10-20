# em

`em` is mya's personal command line assistant containing a variety of utilities.

## Usage

```
$ go install go.pitz.tech/em@latest

$ em -h
NAME:
   em - mya's general purpose command line utilities

USAGE:
   em [options] <command>

COMMANDS:
   analyze      Generate data sets for a variety of integrations.
   auth         Authenticate using common mechanisms.
   encode, enc  Read and write different encodings.
   scaffold     Scaffold out a new project or add onto an existing one.
   storj        Utility scripts for working with storj-specific semantics.
   ulid         Generate or format myago/ulids.
   version      Print the binary version information.

GLOBAL OPTIONS:
   --log_level value   adjust the verbosity of the logs (default: "info") [$LOG_LEVEL]
   --log_format value  configure the format of the logs (default: "console") [$LOG_FORMAT]
   --help, -h          show help (default: false)
```