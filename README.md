# ff-job-manager

基于 AI 生成的 FusionFlex job-manager 工具代码库。生成的代码文件可直接被拷贝或引用到 job-manager 项目中使用。

## 项目结构

```
ff-job-manager/
├── pkg/              # 可被外部项目引用的工具包
│   └── zebualloc/    # ZeBu设备自动分配
├── docs/             # 项目文档
├── specs/            # 工具包的设计文档（实现机制等）
├── go.mod            # Go 模块定义
└── .gitignore
```

## 工具包

### zebualloc — ZeBu设备自动分配

为作业自动选择ZeBu设备，用户声明所需的HalfModule（zs3/zs4）或SubModule（zs5）数量，即可从可用设备列表中返回合适的分配结果。

```go
import "github.com/wjliu/ff-job-manager/pkg/zebualloc"

// zs3/zs4: 分配2个HalfModule
avail := []string{"U0.HM0", "U0.HM1", "U0.HM2", "U0.HM3"}
result, err := zebualloc.Allocate(2, avail, zebualloc.ZS3)
// result: ["U0.HM0", "U0.HM1"]

// zs5: 分配4个SubModule
avail5 := []string{"U0.M0.S0", "U0.M0.S1", "U0.M0.S2", "U0.M0.S3"}
result, err = zebualloc.Allocate(4, avail5, zebualloc.ZS5)
// result: ["U0.M0.S0", "U0.M0.S1", "U0.M0.S2", "U0.M0.S3"]
```

分配规则详见 [specs/zebu-auto-allocation.md](specs/zebu-auto-allocation.md)。

## 使用方式

作为 Go 模块引入：

```bash
go get github.com/wjliu/ff-job-manager
```

## 开发

```bash
# 克隆仓库
git clone https://github.com/wjliu/ff-job-manager.git

# 运行测试
go test ./pkg/...

# 验证模块
go mod verify
```
