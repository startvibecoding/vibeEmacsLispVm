# Examples

This directory contains small examples for embedding and running
vibeEmacsLispVm.

## Embedding

Run the Go embedding example:

```bash
go run ./examples/embedding
```

## CLI Script

Build the CLI and run the sample Lisp script:

```bash
make build
./bin/elispvm -file ./examples/scripts/basic.el
```
