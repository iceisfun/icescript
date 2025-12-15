# 02_interop

This example demonstrates how to inject global variables and functions from the Go host into the Icescript environment.

## Usage

```bash
go run main.go script.ice
```

## How it works
The host application (Go) defines "Config" (a string) and "Callback" (a function) in the symbol table before compilation. It then pre-populates these values in the VM before execution. The script can then use these symbols as if they were global variables.
