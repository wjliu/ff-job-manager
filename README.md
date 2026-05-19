# ff-job-manager

基于 AI 生成的 FusionFlex job-manager 工具代码库。生成的代码文件可直接被拷贝或引用到 job-manager 项目中使用。

## 项目结构

```
ff-job-manager/
├── pkg/       # 可被外部项目引用的工具包
├── docs/      # 项目文档
├── specs/     # 工具包的设计文档（实现机制等）
├── go.mod     # Go 模块定义
└── .gitignore
```

## 使用方式

作为 Go 模块引入：

```bash
go get github.com/wjliu/ff-job-manager
```

在代码中引用工具包：

```go
import "github.com/wjliu/ff-job-manager/pkg/..."
```

## 开发

```bash
# 克隆仓库
git clone https://github.com/wjliu/ff-job-manager.git

# 验证模块
go mod verify
```
