package auxlib

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/parser"
	"github.com/iceisfun/icescript/vm"
	"github.com/redis/go-redis/v9"
)

// ScriptService defines the interface for script management
type ScriptService interface {
	List(ctx context.Context) ([]string, error)
	Load(ctx context.Context, name string) (string, error)
	Save(ctx context.Context, name, content string) error
	Delete(ctx context.Context, name string) error
	Test(ctx context.Context, content string) (*TestResult, error)
}

// ScriptStorage defines the interface for storage backends
type ScriptStorage interface {
	List(ctx context.Context) ([]string, error)
	Load(ctx context.Context, name string) (string, error)
	Save(ctx context.Context, name, content string) error
	Delete(ctx context.Context, name string) error
}

type TestResult struct {
	Output string
	Error  string
}

// Service implements ScriptService using a ScriptStorage backend
type Service struct {
	storage       ScriptStorage
	vmCreateFn    func(*compiler.Bytecode) *vm.VM
	testHarnessFn func(context.Context, *vm.VM) error
}

type Option func(*Service)

// RedisStorage implements ScriptStorage using Redis
type RedisStorage struct {
	client *redis.Client
	prefix string
}

func NewRedisStorage(client *redis.Client, prefix string) *RedisStorage {
	return &RedisStorage{
		client: client,
		prefix: prefix,
	}
}

func (r *RedisStorage) List(ctx context.Context) ([]string, error) {
	var scripts []string
	iter := r.client.Scan(ctx, 0, r.prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		name := strings.TrimPrefix(key, r.prefix)
		scripts = append(scripts, name)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return scripts, nil
}

func (r *RedisStorage) Load(ctx context.Context, name string) (string, error) {
	key := r.prefix + name
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("script not found")
		}
		return "", err
	}
	return val, nil
}

func (r *RedisStorage) Save(ctx context.Context, name, content string) error {
	return r.client.Set(ctx, r.prefix+name, content, 0).Err()
}

func (r *RedisStorage) Delete(ctx context.Context, name string) error {
	return r.client.Del(ctx, r.prefix+name).Err()
}

// NewService creates a new script service with the given storage
func NewService(storage ScriptStorage, opts ...Option) *Service {
	svc := &Service{
		storage: storage,
		vmCreateFn: func(bc *compiler.Bytecode) *vm.VM {
			return vm.New(bc)
		},
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

// NewRedisScriptService helper to maintain backward compatibility or ease of use
// It constructs a RedisStorage and a Service.
// Note: Options now apply to Service, but we might need Redis specific options.
// Actually, the previous design mixed them.
// Let's separate them. But `cmd/editor` uses `NewRedisScriptService` with opts like `WithRedisHost`.
// So we need to parse those options to build the RedisStorage, OR `NewRedisScriptService`
// needs to handle both Service options and Redis config.
// Simpler: `NewRedisScriptService` builds the redis client and storage, then returns NewService.
// BUT `WithRedisHost` etc need to modify... what?
// I'll create a temporary config struct for the builder.

type redisServiceBuilder struct {
	serviceOpts []Option
	redisOpts   *redis.Options
	prefix      string
}

// We need to change the Option type if we want it to support both?
// Previously `Option` was `func(*RedisScriptService)`.
// Now `Service` doesn't have redis client.
// I will keep `NewRedisScriptService` but it will largely be a wrapper that configures things.
// However, `WithRedisHost` expects `*RedisScriptService` (which had `client`).
// Refactoring implies breaking changes or adapters.
// Since I can update `cmd/editor`, I will change the API to be cleaner.
// `NewRedisScriptService` will take `redis.Options`?
// Or I'll just rewrite `cmd/editor/main.go` to construct `RedisStorage` manually.
// That is cleaner: "Lets put Redis behind an interface such that the user can bring-their-own-storage".

// So:
// 1. `NewService(storage ScriptStorage, opts ...Option)`
// 2. Options are strict for Service (VM stuff).
// 3. `RedisStorage` has its own constructor/config.
// 4. `cmd/editor` will instantiate RedisStorage, then NewService.

// Support functions for existing Options compatibility?
// `WithVmCreate` -> `Service` option.
// `WithRedisHost` -> No longer a Service option, but a RedisStorage config.

func WithVmCreate(fn func(*compiler.Bytecode) *vm.VM) Option {
	return func(s *Service) {
		s.vmCreateFn = fn
	}
}

func WithVmTestHarness(fn func(context.Context, *vm.VM) error) Option {
	return func(s *Service) {
		s.testHarnessFn = fn
	}
}

func (s *Service) List(ctx context.Context) ([]string, error) {
	return s.storage.List(ctx)
}

func (s *Service) Load(ctx context.Context, name string) (string, error) {
	return s.storage.Load(ctx, name)
}

func (s *Service) Save(ctx context.Context, name, content string) error {
	return s.storage.Save(ctx, name, content)
}

func (s *Service) Delete(ctx context.Context, name string) error {
	return s.storage.Delete(ctx, name)
}

func (s *Service) Test(ctx context.Context, content string) (*TestResult, error) {
	l := lexer.New(content)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return &TestResult{
			Error: fmt.Sprintf("Parse errors:\n%s", strings.Join(p.Errors(), "\n")),
		}, nil
	}

	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		return &TestResult{
			Error: fmt.Sprintf("Compilation error: %s", err),
		}, nil
	}

	bytecode := c.Bytecode()
	machine := s.vmCreateFn(bytecode)

	var outBuf bytes.Buffer
	machine.SetOutput(&outBuf)

	if s.testHarnessFn != nil {
		if err := s.testHarnessFn(ctx, machine); err != nil {
			return &TestResult{
				Output: outBuf.String(),
				Error:  fmt.Sprintf("Harness error: %s", err),
			}, nil
		}
	}

	err = machine.Run(ctx)
	output := outBuf.String()

	if err != nil {
		return &TestResult{
			Output: output,
			Error:  fmt.Sprintf("Runtime error: %s", err),
		}, nil
	}

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped != nil {
		output += fmt.Sprintf("\nResult: %s", lastPopped.Inspect())
	}

	return &TestResult{
		Output: output,
	}, nil
}
