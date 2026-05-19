## 项目概述

ff-job-manager 是基于 AI 生成的 FusionFlex job-manager 工具代码库，生成的代码可直接被拷贝或引用。

## 项目结构

```
ff-job-manager/
├── pkg/       # 可被外部项目引用的 Go 工具包
├── docs/      # 项目文档（Markdown）
├── specs/     # 工具包设计文档（实现机制等，区别于 docs）
├── go.mod     # Go 模块定义（github.com/wjliu/ff-job-manager）
└── .gitignore
```

### 目录说明

- **pkg/**: 工具包目录，定义 job-manager 项目会使用到的工具代码，可被外部项目引用
- **docs/**: 项目相关文档，如架构说明等
- **specs/**: 工具包的设计文档，主要与工具包的实现机制等设计内容有关

## Go 代码规范

- Go 模块路径: `github.com/wjliu/ff-job-manager`
- 新增工具包放置于 `pkg/` 下，包名使用小写单词，不使用下划线或驼峰
- 导出函数和类型需添加标准 Go doc 注释
- 使用 `go fmt` 格式化代码

## Version Control

This project is managed with Git. The Go module path is `github.com/wjliu/ff-job-manager`.

### Commit Co-Authored-By

Every commit must include a `Co-Authored-By` line indicating the AI model that generated the code. The model name **must match the actual model in use** — do not blindly copy from other environments or sessions.

Before writing the signature, check the current model identity. Examples:
- Using GLM via 火山方舟: `Co-Authored-By: GLM-5.1 <noreply@zhipuai.cn>`
- Using Claude via Anthropic: `Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>`

## Language

Documentation and specs are written in Chinese (中文). Code comments and commit messages may be in either Chinese or English.
