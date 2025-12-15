# Auxlib Package

The `auxlib` package provides helper services for managing and testing Icescript execution. It is designed to be storage-agnostic, allowing you to persist scripts in Redis, local files, or any custom backend.

## Responsibilities

The auxlib package is responsible for:
- Persisting script source code
- Listing, loading, saving, and deleting scripts
- Acting as a bridge between storage and host applications

The auxlib package does NOT:
- Execute scripts
- Manage VM lifecycle
- Enforce security policies
- Provide isolation or sandboxing

## ScriptStorage Interface

To bring your own storage, implement the `ScriptStorage` interface:

```go
type ScriptStorage interface {
    // List returns a list of script names available in storage.
    List(ctx context.Context) ([]string, error)

    // Load retrieves the content of a script by name.
    Load(ctx context.Context, name string) (string, error)

    // Save persists a script's content under the given name.
    Save(ctx context.Context, name, content string) error

    // Delete removes a script from storage.
    Delete(ctx context.Context, name string) error
}
```

## Provided Implementations

### RedisStorage

We provide a example Redis implementation.

**Usage:**

```go
import (
    "github.com/iceisfun/icescript/auxlib"
    "github.com/redis/go-redis/v9"
)

func main() {
    // 1. Initialize Redis Client
    rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

    // 2. Create Storage
    storage := auxlib.NewRedisStorage(rdb, "my-app-prefix:")

    // 3. Create Service
    svc := auxlib.NewService(storage)
}
```

## Implementing Custom Storage

You can easily implement your own storage backend on top of your preferred database (Postgres, MongoDB, Filesystem, etc.).

**Example: In-Memory Map Storage**

```go
package main

import (
    "context"
    "fmt"
    "sync"
)

type MemoryStorage struct {
    mu      sync.RWMutex
    scripts map[string]string
}

func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        scripts: make(map[string]string),
    }
}

func (m *MemoryStorage) List(ctx context.Context) ([]string, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    var names []string
    for k := range m.scripts {
        names = append(names, k)
    }
    return names, nil
}

func (m *MemoryStorage) Load(ctx context.Context, name string) (string, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    content, ok := m.scripts[name]
    if !ok {
        return "", fmt.Errorf("script not found")
    }
    return content, nil
}

func (m *MemoryStorage) Save(ctx context.Context, name, content string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.scripts[name] = content
    return nil
}

func (m *MemoryStorage) Delete(ctx context.Context, name string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.scripts, name)
    return nil
}
```
