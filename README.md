# NAME

em - mya's general purpose command line utilities

# SYNOPSIS

em

```
[--help|-h]
[--log_format]=[value]
[--log_level]=[value]
```

**Usage**:

```
em [options] <command>
```

# GLOBAL OPTIONS

**--help, -h**: show help

**--log_format**="": configure the format of the logs (default: json)

**--log_level**="": adjust the verbosity of the logs (default: info)

# COMMANDS

## admin

Administrative functions for em.

**--help, -h**: show help

### docs

Documentation for em.

**--help, -h**: show help

#### markdown

Prints markdown documentation for em.

**--help, -h**: show help

#### man

Prints the man page for em.

## encode, enc

Read and write different encodings.

> em encode [message]

**--in, -i**="": the input encoding (default: ascii)

**--out, -o**="": the output encoding (default: ascii)

## jenkins

Common operations for working with Jenkins deployments.

### builds

Common operations for working with Jenkins builds.

#### analyze

Analyze builds in a Jenkins instance

**--index_database_dsn**="": specify the connection string for database (default: file:db.sqlite)

**--jenkins_base_url**="": specify the base url of the jenkins instance we're indexing

**--jenkins_job**="": provide an initial list of jobs to analyze

## kubernetes, kube

Common operations for working with Kubernetes resources.

## oidc

Common operations for working with OIDC providers.

### auth

Authenticate with an OIDC provider.

**--client_id**="": the client_id associated with this service

**--client_secret**="": the client_secret associated with this service

**--issuer_certificate_authority**="": path pointing to a file containing the certificate authority data for the server

**--issuer_server_url**="": the address of the server where user authentication is performed

**--redirect_url**="": the redirect_url used by this service to obtain a token

**--scopes**="": specify the scopes that this authorization requires

## project

Common operations for working with projects.

### scaffold

Scaffold out a new project or add onto an existing one.

    em project scaffold [options] <name>
       em project scaffold features    # will output a list of features and aliases
       em project scaffold --mkdir --license mpl --features init <name>
       em project scaffold --mkdir --license mpl --features init --features bin <name>

**--features**="": specify the features to generate

**--license**="": specify which license should be applied to the project (default: agpl3)

**--mkdir**: specify if we should make the target project directory

## storj

Common operations for working with Storj resources.

### auth

Authenticate with a Storj OIDC provider.

**--client_id**="": the client_id associated with this service

**--client_secret**="": the client_secret associated with this service

**--issuer_certificate_authority**="": path pointing to a file containing the certificate authority data for the server

**--issuer_server_url**="": the address of the server where user authentication is performed

**--redirect_url**="": the redirect_url used by this service to obtain a token

**--scopes**="": specify the scopes that this authorization requires

### uuid

Format storj-specific UUID.

**--out, -o**="": specify the output format (string or bytes) (default: string)

#### format

Swap between different formats of the UUID (string and bytes)

**--in, -i**="": specify the input format (string or bytes) (default: string)

**--out, -o**="": specify the output format (string or bytes) (default: bytes)

## ulid

Generate or format myago/ulids.

**--out, -o**="": specify the output format (string, bytes) (default: string)

**--size**="": specify the size of the ulid being generated (default: 256)

### format

Parse and format provided myago/ulids.

**--in, -i**="": specify the input format (string, bytes) (default: string)

**--out, -o**="": specify the output format (json, string, bytes) (default: json)

## version

Print the binary version information.

> em version
