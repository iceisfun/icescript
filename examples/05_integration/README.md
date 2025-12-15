# 05_integration

This example demonstrates deep integration between Go and Icescript, including shared global state modification.

## Usage

```bash
go run main.go script.ice
```

## Description
The host provides `version` and `config` globals to the VM. The script modifies these globals, and the host inspects the final state after execution.
