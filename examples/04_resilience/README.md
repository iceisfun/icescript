# 04_resilience

This example demonstrates the VM's ability to handle context cancellation (timeouts) to prevent runaway scripts.

## Usage

```bash
go run main.go script.ice
```

## Description
The script enters an infinite loop. The host application sets a context timeout, ensuring the VM aborts execution gracefully.
