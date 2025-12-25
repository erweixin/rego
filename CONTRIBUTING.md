# Contributing to Rego

Thank you for your interest in Rego! We welcome contributions of all kinds, including bug reports, documentation improvements, and code contributions.

## How to Contribute

1. **Fork** the repository.
2. **Create your feature branch** (`git checkout -b feature/AmazingFeature`).
3. **Commit your changes** (`git commit -m 'Add some AmazingFeature'`).
4. **Push to the branch** (`git push origin feature/AmazingFeature`).
5. **Open a Pull Request**.

## Core Principles

Rego's core goal is to **bring the React Hooks development experience to the TUI world**. When contributing code, please follow these principles:

- **Hooks First**: State and side effects should be managed through Hooks.
- **Explicit Over Implicit**: Use explicit keys to identify state.
- **Type Safety**: Leverage Go's generics fully.
- **Clean UI API**: Keep the Node and Style APIs simple and composable.

## Development Setup

```bash
# Clone the repository
git clone https://github.com/erweixin/rego.git
cd rego

# Run tests
go test ./...

# Run examples
cd examples/gallery
go run .
```

## Testing

Before submitting a PR, please ensure all tests pass:

```bash
go test ./...
```

If you add new features, please add corresponding test cases. You can use the `testing` package for component tests and snapshot tests.

## Code Style

Please follow standard Go code style. Run `go fmt` before committing.

## Pull Request Guidelines

- Keep PRs focused on a single feature or bug fix
- Include tests for new functionality
- Update documentation if needed
- Follow existing code patterns and conventions

## Reporting Issues

When reporting bugs, please include:
- Go version
- Operating system
- Steps to reproduce
- Expected vs actual behavior

## Code of Conduct

Please be professional and friendly. We are committed to providing a harassment-free environment for everyone.

---

# 贡献指南

感谢你对 Rego 的关注！我们欢迎任何形式的贡献，包括提交 Bug、改进文档或贡献代码。

## 如何贡献

1. **Fork** 本仓库。
2. **创建你的特性分支** (`git checkout -b feature/AmazingFeature`)。
3. **提交你的更改** (`git commit -m 'Add some AmazingFeature'`)。
4. **推送到分支** (`git push origin feature/AmazingFeature`)。
5. **开启一个 Pull Request**。

## 核心理念

Rego 的核心目标是**将 React Hooks 的开发体验带入 TUI 世界**。在贡献代码时，请遵循以下原则：

- **Hooks 优先**：状态和副作用应通过 Hooks 管理。
- **显式优于隐式**：使用显式的 key 来标识状态。
- **类型安全**：充分利用 Go 的泛型。
- **简洁的 UI API**：保持节点 (Node) 和样式 (Style) 的 API 简洁且易于组合。

## 测试

在提交 PR 之前，请确保所有测试都能通过：

```bash
go test ./...
```

如果你添加了新功能，请务必添加相应的测试用例。

## 代码风格

请遵循标准的 Go 代码风格，建议运行 `go fmt`。

## 行为准则

请保持专业和友善。我们致力于为每位参与者提供一个无骚扰的环境。
