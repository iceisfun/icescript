# 08_error_handling

This example demonstrates how runtime errors (including `panic()` calls and stack traces) are reported to the host application.

## Usage

```bash
go run main.go script.ice
```

## Description
The script is designed to fail. The host application captures the error returned by `machine.Run()` and prints it, showing the custom error message and stack trace.
