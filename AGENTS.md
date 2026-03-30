# DeerFlow Go - Agent Guidelines

## Project Overview

Go implementation of DeerFlow backend - a "super agent harness" using standard Go patterns.
Module: `github.com/user/deer-flow-go`
Go version: >= 1.25

## AI 编程规范

### 技术方案原则

- 技术方案核心内容: 技术栈、代码规范、架构规范
- **严格遵守技术栈的版本信息**: 如依赖的技术栈版本超出模型训练截止版本，**必须调用搜索工具确认**新版本的能力、API等
- **输出规范**: 技术方案探讨阶段**不要输出具体代码**，只输出技术方案
- **错误处理规范**: 严格遵守 Go 语言最佳实践

### Go 版本与标准库

- Go version >= 1.25
- 使用新版标准库:
  - `math/rand/v2`
  - `encoding/json/v2`

### 代码质量要求

- Time Complexity 优化
- 无注释原则: 代码自文档化，通过命名表达意图
- 单元测试覆盖
- 项目结构规范
- 代码整洁度

### 提交要求

提交完整工程，包括所需的所有文件:
- 代码文件
- 新增的测试用例
- README 等文档

## 工具

```sh
# golangci-lint - https://golangci-lint.run/docs
brew install golangci-lint

# gofumpt - 格式化工具
go install mvdan.cc/gofumpt@latest
gofumpt -l -w .

# modernize - 代码现代化工具
go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -test ./...
```

## 依赖包

| 包 | 状态 | 说明 |
|---|---|---|
| `github.com/openai/openai-go/v3` | ✅ 使用 | OpenAI 官方 SDK |
| `github.com/modelcontextprotocol/go-sdk` | ✅ 使用 | MCP 官方 SDK |
| `langchaingo` | ❌ 不使用 | 不支持 langchain 1.0 版本 |

## Build/Test Commands

```bash
make build          # Build binary to bin/server
make run            # Run server (go run ./cmd/server)
make test           # Run all tests: go test -race ./... -v
make fmt            # Format code: go fmt ./...
make lint           # Run golangci-lint run ./...
make vet            # Run go vet ./...
make clean          # Remove bin/ and .deer-flow/
```

### Running Single Tests

```bash
go test -race -v ./internal/handler -run TestModelsHandler_List
go test -race -v ./internal/store -run TestThreadStore_Create
go test -race -v ./path/to/package -run TestFunctionName
```

## Project Structure

```
deer-flow-go/
├── cmd/server/         # Entry point (main.go)
├── internal/
│   ├── config/         # Configuration (config.go, types.go)
│   ├── handler/        # HTTP handlers + routes.go
│   ├── logic/          # Business logic layer
│   ├── agent/          # Agent orchestration (engine.go)
│   ├── llm/            # LLM abstractions + providers/
│   ├── store/          # Data storage (thread.go, run.go, memory.go)
│   ├── model/          # Data models (message.go, thread.go, run.go)
│   ├── skills/         # Skills loader
│   └── mcp/            # MCP integration
├── pkg/utils/          # Shared utilities
├── config.yaml         # Runtime config (copy from config.example.yaml)
└── Makefile
```

## Code Style Guidelines

### Imports

Group imports in order:
1. Standard library (fmt, net/http, os, etc.)
2. External dependencies (github.com/google/uuid, etc.)
3. Project imports (github.com/user/deer-flow-go/...)

```go
import (
    "encoding/json"
    "net/http"

    "github.com/google/uuid"

    "github.com/user/deer-flow-go/internal/model"
    "github.com/user/deer-flow-go/internal/store"
)
```

### Formatting

- Use `go fmt` before committing
- No trailing whitespace
- Tabs for indentation (Go standard)

### Types and Structs

- Use `json` tags for API response models
- Use `mapstructure` tags for config structs
- Private fields by default; expose via methods
- Pointer receivers for methods that modify state

```go
type Thread struct {
    ThreadID  string         `json:"thread_id"`
    CreatedAt time.Time      `json:"created_at"`
    Metadata  map[string]any `json:"metadata,omitempty"`
}

type ModelConfig struct {
    Name        string `mapstructure:"name"`
    APIKey      string `mapstructure:"api_key"`
}
```

### Naming Conventions

- **Handlers**: `*Handler` (ThreadsHandler, RunsHandler)
- **Stores**: `*Store` (ThreadStore, RunStore)
- **Constructors**: `New*` (NewThreadsHandler, NewThreadStore)
- **Interfaces**: single-method interfaces named by action (Provider)
- **Errors**: `Err*` prefix (ErrThreadNotFound, ErrRunNotFound)
- **Test functions**: `Test*_*` pattern (TestThreadStore_Create)

### Error Handling

1. **Sentinel errors** for expected failures:
```go
var ErrThreadNotFound = errors.New("thread not found")
```

2. **Wrap errors** with context:
```go
return nil, fmt.Errorf("failed to get provider: %w", err)
```

3. **HTTP handlers** use helpers:
```go
writeError(w, http.StatusNotFound, "thread not found", "")
writeError(w, http.StatusInternalServerError, "failed to create", err.Error())
```

4. **Must* functions** panic on failure (for startup):
```go
func MustLoad(path string) *Config {
    cfg, err := Load(path)
    if err != nil {
        panic(fmt.Sprintf("failed to load config: %v", err))
    }
    return cfg
}
```

### HTTP Patterns

- Use Go 1.22+ pattern routing with `http.ServeMux`
- Path parameters via `r.PathValue("id")`
- Route registration in `routes.go`

```go
mux.HandleFunc("GET /api/langgraph/threads/{id}", r.threads.Get)
mux.HandleFunc("POST /api/langgraph/threads/{id}/runs", r.runs.Create)

func (h *ThreadsHandler) Get(w http.ResponseWriter, r *http.Request) {
    threadID := r.PathValue("id")
    if threadID == "" {
        writeError(w, http.StatusBadRequest, "thread_id is required", "")
        return
    }
    // ...
}
```

### Testing Conventions

**极简原则**:

- ✅ 仅使用 Go 标准库 `testing` 包
- ✅ `xxx_test.go` 中只测试 `xxx.go` 中的函数
- ✅ 使用 option 设计模式进行配置/变量初始化，不需要测试
- ❌ 不使用接口注入测试替身（Test Double）
- ❌ 不使用第三方 mock 框架（gomock、testify/mock）
- ❌ 不使用断言库，使用 `if` 判断

**标准模式**:

- Use `t.TempDir()` for test directories (auto-cleanup)
- Use `httptest.NewRequest` + `httptest.NewRecorder` for HTTP tests
- Set path values with `req.SetPathValue("name", "value")`
- Table-driven tests for multiple cases

```go
func TestThreadStore_Create(t *testing.T) {
    tmpDir := t.TempDir()
    store := NewThreadStore(tmpDir)

    thread, err := store.Create()
    if err != nil {
        t.Fatalf("Create() error = %v", err)
    }
    // assertions...
}

func TestModelsHandler_Get_Found(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/api/models/gpt-4o", nil)
    req.SetPathValue("name", "gpt-4o")
    rec := httptest.NewRecorder()

    handler.Get(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("Get() status = %d, want %d", rec.Code, http.StatusOK)
    }
}
```

## Configuration

Copy `config.example.yaml` to `config.yaml`:
```yaml
server:
  name: deer-flow-go
  host: 0.0.0.0
  port: 8001

models:
  - name: gpt-4o
    display_name: GPT-4o
    use: openai
    api_key: $OPENAI_API_KEY

storage:
  data_dir: .deer-flow
```

Environment variables referenced with `$VAR_NAME` syntax.

## Key Dependencies

- `github.com/spf13/viper` - Configuration
- `github.com/google/uuid` - UUID generation
- `github.com/joho/godotenv` - .env loading
- Standard `net/http` for HTTP server (no framework)

## Important Notes

- No comments in production code (self-documenting via naming)
- Atomic file writes: write to `.tmp`, then rename
- Thread-safe stores with `sync.RWMutex`
- Environment variable config via `godotenv.Load()`
- Fish shell used for command execution