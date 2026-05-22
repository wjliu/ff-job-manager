# 自动分配设备

## 需求描述

该功能旨在为作业自动选择Palladium设备，用户只需要在提交作业时声明需要 Domain 的数量，该工具便可以在可用的 Domain 列表中选择出适合的设备返回。

## 背景知识

Palladium是Candence的一款用于做芯片验证的硬件仿真系统，其中涉及的概念如下：
- Rack：每个Rack最多包含3个Cluster，Rack的编号从0开始递增
- Cluster：每个Cluster最多包含6个Logic Drawer（以下简称LD），Cluster的编号从0开始递增，在多个Rack之间Cluster的编号是连续的
- LD：每个LD有8个Domain，LD的编号从0开始递增，在多个Rack和Cluster之间LD的编号是连续的
- Domain：可以使用的最小设备单元，Domain的编号在LD内部从0开始递增，在多个LD之间Domain的编号不连续。如果每个用户使用一个Domain，则一个Cluster可以支持48个用户同时使用，但是实际使用场景中几乎没有只有一个Domain的情况

## 输入

- Palladium的机型，比如：Palladium Z1、Palladium Z2
- 需要分配的Domain的数量
- 可用的Domain数组，格式如：[0.0, 0.1, 0.2， 0.3]，其中点号分隔的前半部分为LD的编号，点号后半部分为LD内的Domain编号。编号规则可以参考背景知识部分

## 输出

- 分配的Domain数组，格式如：[0.0, 0.1, 0.2， 0.3]
- 分配的Domain所在的Rack的编号列表，格式如：[0, 1, 2]

## 分配规则

在实际分配时应遵守如下规则：
1. 分配的Domain必须是连续的，当请求的Domain数量小于等于8时，不允许跨LD
2. 分配的LD必须是连续的，当请求的LD数量小于等于6时，不允许跨Cluster；当请求的LD数量大于6时，必须从某个Cluster的0号LD开始分配
3. 分配时在满足规则的情况下，应尽量减少资源碎片问题

## 分配机制

### 拓扑常量

| 参数 | 值 | 说明 |
|------|---|------|
| domainsPerLD | 8 | 每个LD包含8个Domain |
| ldPerCluster | 6 | 每个Cluster包含6个LD |
| clustersPerRack | 3 | 每个Rack包含3个Cluster |
| ldPerRack | 18 | 每个Rack包含18个LD |

从LD编号推算：
- Cluster索引 = ldIndex / 6
- Rack索引 = ldIndex / 18
- Cluster内首LD索引 = clusterIndex * 6

### 整体流程

```
输入: requestedCount, available[], machineType
  │
  ├─ 1. 参数校验（数量 > 0, 列表非空）
  │
  ├─ 2. 解析可用列表，按LD分组（buildLDMap）
  │     解析 "LD.Domain" → 按LD分组 → 排序
  │
  ├─ 3. 计算所需LD数和余数
  │     neededLDs = ceil(requestedCount / 8)
  │     remainder = requestedCount % 8（余数为0时等于8）
  │
  ├─ 4. 执行分配
  │     ├─ neededLDs == 1: 单LD内分配（allocateWithinLD）
  │     └─ neededLDs > 1:  跨LD分配（allocateAcrossLDs）
  │
  ├─ 5. 转换为Domain名称列表
  │
  ├─ 6. 截取输出至实际请求数量（非倍数场景）
  │
  └─ 7. 计算涉及的Rack编号列表（去重排序）
```

### 步骤详解

#### 步骤1: 解析可用列表

解析格式为 `LD编号.Domain编号` 的字符串，按LD分组，每个LD内Domain排序。

正则匹配：`^(\d+)\.(\d+)$`

#### 步骤2: 单LD内分配（neededLDs == 1，即 <= 8 Domain）

在单个LD内寻找连续可用Domain，按碎片最小化优先级：
1. **全局优先从D0开始**：遍历所有LD，优先选择D0起有连续可用Domain的LD
2. **其次末尾对齐D7**：选择最后一个可用Domain为D7且从 `8-count` 开始连续的LD
3. **再次任意连续段**：选择任意位置有连续可用Domain的LD

#### 步骤3: 跨LD分配（neededLDs > 1，即 > 8 Domain）

寻找连续LD序列，需满足：
- 前面 `neededLDs-1` 个LD必须完整（8个Domain全部可用）
- 最后1个LD（若有余数）必须从D0开始有 `remainder` 个连续可用Domain
- 最后1个LD若余数为8，也需要完整

Cluster约束：
- **neededLDs <= 6**：所有LD必须在同一个Cluster内
- **neededLDs > 6**：起始LD必须是某Cluster的0号LD（即 `ldIndex % 6 == 0`）

碎片最小化：优先从Cluster的LD0开始分配。

#### 步骤4: 截取输出

当请求数量不是8的倍数时，`ldDomainsToNames` 会输出最后一个LD的全部8个Domain，需截取至实际请求数量 `requestedCount`。

例如：申请10个Domain → 分配2个LD → ldDomainsToNames输出16个 → 截取前10个。

#### 步骤5: Rack计算

从分配的LD列表计算涉及的Rack编号：`rackIndex = ldIndex / 18`，去重排序后返回。

### 分配示例

#### 示例1: 申请4个Domain（单LD内）

可用: LD 0(D0-D7), LD 1(D0-D7)

```
neededLDs = 1, 在单LD内分配
碎片优先: 从D0开始
结果: [0.0, 0.1, 0.2, 0.3], Racks=[0]
```

#### 示例2: 申请8个Domain（1个完整LD）

可用: LD 0(D0-D7), LD 1(D0-D7)

```
neededLDs = 1, 在单LD内分配
结果: [0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7], Racks=[0]
```

#### 示例3: 申请9个Domain（1完整LD + 1 Domain）

可用: LD 0(D0-D7), LD 1(D0-D7)

```
neededLDs = 2, remainder = 1
LD 0完整 + LD 1.D0
结果: [0.0~0.7, 1.0], Racks=[0]
```

#### 示例4: 申请48个Domain（1个完整Cluster）

可用: LD 0-5 (D0-D7)

```
neededLDs = 6, <=6, 在单Cluster内
结果: [0.0~5.7], Racks=[0]
```

#### 示例5: 申请49个Domain（跨Cluster，>6 LD）

可用: LD 0-17 (D0-D7)

```
neededLDs = 7, >6, 起始必须为Cluster LD0
从LD 0(Cluster 0 LD0)开始: LD 0-5(完整) + LD 6(1 Domain)
结果: [0.0~5.7, 6.0], Racks=[0]
```

#### 示例6: 部分LD的D0不可用（跨LD失败）

可用: LD 0(D0-D7), LD 1(D2-D7, D0-D1不可用)

```
申请10 Domain = 1完整LD + 2 Domain
LD 0完整, 但LD 1的D0-D1不可用, 最后LD必须从D0开始
结果: 分配失败
```

#### 示例7: 碎片优先从D0起始

可用: LD 0(D4-D7), LD 1(D0-D7)

```
申请4 Domain, neededLDs = 1
全局优先从D0: LD 0无D0, LD 1有D0
结果: [1.0, 1.1, 1.2, 1.3], Racks=[0]
```

#### 示例8: 碎片优先末尾对齐D7

可用: LD 0(D4-D7)

```
申请4 Domain, neededLDs = 1
无D0起始的LD, 末尾对齐D7: D4-D7
结果: [0.4, 0.5, 0.6, 0.7], Racks=[0]
```

#### 示例9: 非倍数请求数量（截取输出）

可用: LD 0(D0-D7)

```
申请3 Domain（不是8的倍数）
neededLDs = 1, 分配D0-D7(8个)
截取: 前3个 → [0.0, 0.1, 0.2]
```

#### 示例10: >6 LD从Cluster LD0开始

可用: LD 1-12(D0-D7), LD 0不可用

```
申请56 Domain = 7 LD, >6
LD 1不是Cluster LD0, 跳过
LD 6是Cluster 1 LD0, 从LD 6开始: LD 6-12
结果: [6.0~12.7], Racks=[0]
```

#### 示例11: 跨Rack分配

可用: LD 0-35(D0-D7)

```
申请152 Domain = 19 LD, >6, 从LD 0(Cluster 0 LD0)开始
LD 0-18, Rack 0(LD 0-17) + Rack 1(LD 18)
结果: [0.0~18.7], Racks=[0, 1]
```

### 错误场景

| 场景 | 错误信息 |
|-----|---------|
| 请求数量 <= 0 | `requested count must be positive, got {n}` |
| 可用设备列表为空 | `no available devices` |
| 单LD内无足够连续Domain | `cannot allocate {n} consecutive domains in any LD` |
| <=6 LD无法在单Cluster内分配 | `cannot allocate {n} consecutive LDs within a cluster` |
| >6 LD无法从Cluster LD0开始分配 | `cannot allocate {n} consecutive LDs starting from cluster LD0` |
| Domain名称格式无效 | `invalid Domain name: {name}, expected format like 0.0` |
| Domain索引超出范围 | `invalid Domain index: {n} in {name}, must be 0-7` |
| 未知机型 | `unknown machine type: {s}` |
