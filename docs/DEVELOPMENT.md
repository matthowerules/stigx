# Development Setup Guide

## Prerequisites

### System Dependencies
- Go 1.24+ 
- Node.js 18+ and npm
- Git

### Linux-specific Dependencies (Required for Wails)
```bash
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev
```

### Optional Dependencies
```bash
# For Windows builds on Linux
sudo apt install nsis
```

## Quick Start

### 1. Clone and Setup
```bash
git clone <repository-url>
cd stigx
go mod download
```

### 2. Frontend Setup
```bash
cd frontend
npm install
```

### 3. Development Builds

#### CLI Development
```bash
# Run CLI directly
go run cmd/cli/main.go --help

# Test specific commands
go run cmd/cli/main.go convert --help
go run cmd/cli/main.go validate --help
```

#### GUI Development
```bash
# Run in development mode with hot reload
wails dev

# Or run manually
cd frontend && npm run dev &
go run cmd/stigx/main.go
```

### 4. Production Builds

#### CLI Binary
```bash
# Build for current platform
go build -o build/stigx-cli cmd/cli/main.go

# Cross-compile for other platforms
GOOS=windows GOARCH=amd64 go build -o build/stigx-cli.exe cmd/cli/main.go
GOOS=darwin GOARCH=amd64 go build -o build/stigx-cli-macos cmd/cli/main.go
```

#### GUI Application
```bash
# Build for current platform
wails build

# Build for production with optimizations
wails build -clean -upx -s

# Cross-platform builds
wails build -platform windows/amd64
wails build -platform darwin/amd64
wails build -platform linux/amd64
```

## Project Structure

```
stigx/
├── cmd/
│   ├── cli/          # CLI application entry point
│   └── stigx/        # GUI application entry point (Wails)
├── internal/
│   ├── app/          # Application logic
│   ├── cli/          # CLI command implementations
│   ├── models/       # Data models (CKL, CKLb, etc.)
│   ├── parsers/      # File parsing logic
│   └── converters/   # Format conversion logic
├── pkg/              # Public packages
├── frontend/         # Wails frontend (React + TypeScript)
├── test/             # Tests
├── docs/             # Documentation
├── scripts/          # Build and utility scripts
└── references/       # Sample files and documentation
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/models
```

## Code Style

```bash
# Format code
go fmt ./...

# Run linter (install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
golangci-lint run

# Update dependencies
go mod tidy
```

## Debugging

### VS Code
Create `.vscode/launch.json`:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "CLI Debug",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/cli/main.go",
            "args": ["convert", "test.ckl", "test.cklb"]
        },
        {
            "name": "GUI Debug", 
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/stigx/main.go"
        }
    ]
}
```

### Wails Debug Mode
```bash
wails dev -devtools
```

## Contributing

1. Create feature branch from `main`
2. Implement changes with tests
3. Run `go fmt`, `go vet`, and `golangci-lint run`
4. Test CLI and GUI functionality
5. Update documentation if needed
6. Submit pull request

## Common Issues

### Wails Build Failures
- Ensure all system dependencies are installed
- Clear build cache: `wails build -clean`
- Check frontend builds independently: `cd frontend && npm run build`

### Import Errors
- Run `go mod tidy` after adding dependencies
- Ensure GOPATH and module setup is correct

### Frontend Issues
- Clear node_modules: `rm -rf frontend/node_modules && cd frontend && npm install`
- Check Node.js version compatibility