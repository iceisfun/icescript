# 06_invoke

This example demonstrates how to invoke specific Icescript functions from Go.

## Usage

```bash
go run main.go script.ice
```

## Description
After running the script once to initialize definitions, the host looks up global functions (`onTick`, `add`, `slow`) and calls them directly with arguments using `machine.Invoke`.
