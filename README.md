# du-split

> 项目初始化：该仓库已完成初始化，包含基础 Go 项目结构与说明。

A lightweight Go utility for splitting disk usage or directory contents into balanced groups.

## Project structure

```text
du-split/
├── cmd/
│   └── du-split/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── app.go
│   ├── scanner/
│   │   └── scanner.go
│   └── report/
│       └── report.go
├── pkg/
│   └── util/
│       └── file.go
├── go.mod
├── README.md
└── .gitignore
```

## Getting started

```bash
go run ./cmd/du-split
```
