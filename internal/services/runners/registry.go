package runners

import (
	"context"
	"fmt"
	"sync"

	"github.com/griffnb/techboss-ai-go/internal/services/runners/importing"
)

type Runner interface {
	Run(ctx context.Context, args ...string) error
}

var (
	registry = make(map[string]Runner)
	mx       sync.RWMutex
)

func Register(name string, runner Runner) {
	mx.Lock()
	defer mx.Unlock()
	fmt.Printf("Registering %s\n", name)
	registry[name] = runner
}

func Get(name string) Runner {
	mx.RLock()
	defer mx.RUnlock()
	runner, ok := registry[name]
	if !ok {
		return nil
	}
	return runner
}

func init() {
	Register("categories", &importing.CategoryImportRunner{})
	Register("tools", &importing.ToolImportRunner{})
}
