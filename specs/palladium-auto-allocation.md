# 自动分配设备

## 需求描述

该功能旨在为作业自动选择Palladium设备，用户只需要在提交作业时声明需要 Domain 的数量，该工具便可以在可用的 Domain 列表中选择出适合的设备返回。

## 背景知识

Palladium是Candence的一款用于做芯片验证的硬件仿真系统，其中涉及的概念如下：
- Rack：每个Rack最多包含3个Cluster，Rack的编号从0开始递增
- Cluster：每个Cluster最多包含6个Logic Drawer（以下简称LD），Cluster的编号从0开始递增，在多个Rack之间Cluster的编号是连续的
- LD：每个LD有8个Domain，LD的编号从0开始递增，在多个Rack和Cluster之间LD的编号是连续的
- Domain：可以使用的最小设备单元，Domain的编号在LD内部从0开始递增，在多个LD之间Domain的编号不连续。如果每个用户使用一个Domain，则一个Cluster可以支持48个用户同时使用，但是实际使用场景中几乎没有只有一个Domain的情况
- TPod：挂接外设的接口，每个Rack内都包含一定数量的TPod，没有明确的数量上限，TPod的编号在Rack内从0开始递增，不一定每个TPod都会挂接外设

## 输入

- Palladium的机型，比如：Palladium Z1、Palladium Z2。
- 需要分配的Domain的数量。
- 可用的Domain数组，格式如：["0.0","0.1","0.2"，"0.3"]，其中点号分隔的前半部分为LD的编号，点号后半部分为LD内的Domain编号。编号规则可以参考背景知识部分。
- 需要分配的TPod数组，可选，格式如：[{"USB-HDSB", 1}, {"PCI", 2}]，其中数组中每个元素包含两个字段：一个字段为TPod的外设类型（Type），这个没有固定值，是自定义的。另一个字段是需要的该外设类型的数量（Number）。
- 可用的TPod信息（map），格式如：{0: {"USB-HDB": [0,1,2], "PCI": [3,4]}, 1: {"USB-HDB": [2,3,4]}}，其中map的key为RackId，value为TPod的外设类型（ExtType）和对应的TPodId数组的映射。

## 输出

- 分配的Domain数组，格式如：[0.0, 0.1, 0.2， 0.3]
- 分配的Domain所在的Rack的编号列表，格式如：[0, 1, 2]
- 分配的TPods数组，格式如：[{0, 1, "USB-HDSB"}, {0, 2, "USB-HDSB"}]，其中数组中每个元素包含三个字段。第一个字段为Rack的编号（RackId），第二个字段为TPod编号（TPodId），第三个字段为外设类型（ExtType）。

## 分配规则

在实际分配时应遵守如下规则：
1. 分配的Domain必须是连续的，当请求的Domain数量小于等于8时，不允许跨LD
2. 分配的LD必须是连续的，当请求的LD数量小于等于6时，不允许跨Cluster；当请求的LD数量大于6时，必须从某个Cluster的内的第一个LD开始分配；当请求的LD数量大于18，导致跨Rack时，必须从某个Rack的第一个Cluster的第一个LD开始分配
3. 分配时在满足规则的情况下，应尽量减少资源碎片问题
4. 当输入了需要分配的TPod数组时，则必须在满足Domain分配的Rack内同时满足TPod的分配才可以，且TPod的分配不能跨Rack，即便Domain出现了跨Rack的情况，TPod的分配也必须在仅一个Rack中满足才行
5. 如果在已经成功分配Domain的Rack列表中不能满足TPod的分配，则应该继续向后探查是否后续仍有可以满足Domain成功分配也可能满足TPod分配需求的Rack列表，而不是遇到一次不满足TPod失败就结束分配，应该尝试直到所有Rack全部检查完

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
输入: requestedCount, availableDomains[], machineType, tpodReqs[]?, availableTPods? 
  │
  ├─ 1. 参数校验
  │     数量 > 0, Domain列表非空
  │     若有TPod需求: 校验需求有效性, 可用TPod列表非空（validateTPodReqs）
  │
  ├─ 2. 解析可用Domain列表，按LD分组（buildLDMap）
  │     解析 "LD.Domain" → 按LD分组 → Domain编号排序
  │
  ├─ 3. 计算所需LD数和余数
  │     neededLDs = ceil(requestedCount / 8)
  │     remainder = requestedCount % 8（余数为0时等于8）
  │
  ├─ 4. 执行分配（流式：边遍历边检查，满足即返回）
  │     ├─ neededLDs == 1: 单LD内分配（allocateWithinLD）
  │     │     按3个优先级遍历LD，对每个候选即时检查TPod
  │     │     Domain+TPod都满足 → 立即返回，不继续遍历
  │     │     TPod不满足 → 跳过当前LD，继续下一个（规则5）
  │     └─ neededLDs > 1:  跨LD分配（allocateAcrossLDs）
  │           遍历排序后的起始LD，对每个成功的序列即时检查TPod
  │           Domain+TPod都满足 → 立即返回
  │           TPod不满足 → 继续下一个起始LD（规则5）
  │
  ├─ 5. 转换为Domain名称列表（ldDomainsToNames）
  │
  ├─ 6. 截取输出至实际请求数量（非倍数场景）
  │
  └─ 7. 计算涉及的Rack编号列表（computeRacks，去重排序）
```

### 步骤详解

#### 步骤1: 参数校验

- Domain请求数量必须 > 0
- 可用Domain列表非空
- 若有TPod需求（`tpodReqs` 非空）:
  - 每个需求的 `ExtType` 不能为空字符串
  - 每个需求的 `Number` 必须 > 0
  - `availableTPods` 不能为空

#### 步骤2: 解析可用列表

解析格式为 `LD编号.Domain编号` 的字符串，按LD分组，每个LD内Domain排序。

正则匹配：`^(\d+)\.(\d+)$`

#### 步骤3: 执行分配（流式边遍历边检查）

采用**流式处理**：在遍历候选的过程中直接检查 TPod，Domain + TPod 都满足即立即返回。这样避免了"先收集全部候选、再逐一筛选"带来的无条件遍历开销。无 TPod 需求时行为与原有"首次匹配即返回"一致。

##### 3a. 单LD内分配（allocateWithinLD）

按3个优先级遍历 LD，每个候选即时做 TPod 检查。**同一LD只出现在最高匹配优先级中**（去重）：

| 优先级 | 条件 | 说明 |
|-------|------|------|
| 0（最高） | D0 起始 + 连续 count 个 | 全局遍历，优先从D0开始 |
| 1 | 末尾对齐 D7 + 连续 count 个 | D7 为最后一个可用Domain |
| 2（最低） | 任意位置连续 count 个 | 每个LD只取第一个连续段 |

处理流程：
1. 遍历当前优先级的 LD 列表
2. 判断 Domain 连续可用
3. 若无 TPod 需求 → 立即返回该 LD
4. 若有 TPod 需求 → 调用 `allocateTPods` 检查该 LD 所在 Rack
   - 满足 → 立即返回 Domain + TPod
   - 不满足 → 跳过，继续下一个 LD（**规则5**）

##### 3b. 跨LD分配（allocateAcrossLDs）

寻找连续LD序列，需满足：
- 前面 `neededLDs-1` 个LD必须完整（8个Domain全部可用）
- 最后1个LD（若有余数）必须从D0开始有 `remainder` 个连续可用Domain
- 最后1个LD若余数为8，也需要完整

Cluster约束：
- **neededLDs <= 6**：所有LD必须在同一个Cluster内
- **6 < neededLDs <= 18**：起始LD必须是某Cluster的0号LD（即 `ldIndex % 6 == 0`）
- **neededLDs > 18**（跨Rack）：起始LD必须是某Rack的第一个Cluster的第一个LD（即 `ldIndex % 18 == 0`）

候选起始LD按优先级排序（碎片优先：Rack LD0 > Cluster LD0 > 其他），**流式遍历**：每个成功的 `tryAllocateLDs` 结果即时检查 TPod，满足即返回，不继续遍历。

#### 步骤4: TPod 分配（allocateTPods / allocateTPodsInRack）

在流式分配中，每个 Domain 候选产生后即时调用 TPod 分配。从候选的 Rack 列表中找一个能**独立**满足所有 TPod 需求的 Rack：

1. 遍历 Rack 列表
2. 对每个 Rack，检查 `availableTPods[rackId]` 中是否每个 ExtType 都有足够数量的 TPodId
3. 找到第一个满足的 Rack，取每种 ExtType 的前 N 个 TPodId 返回
4. 所有 Rack 都不满足则返回错误（触发规则5继续下一个 Domain 候选）

**关键约束（规则4）**：TPod 不能跨 Rack——即便 Domain 跨越多个 Rack，所有 TPod 需求必须在单个 Rack 内满足。

#### 步骤5: 截取输出

当请求数量不是8的倍数时，`ldDomainsToNames` 会输出最后一个LD的全部8个Domain，需截取至实际请求数量 `requestedCount`。

例如：申请10个Domain → 分配2个LD → ldDomainsToNames输出16个 → 截取前10个。

#### 步骤6: Rack计算

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

#### 示例11: 跨Rack分配（>18 LD，从Rack LD0开始）

可用: LD 0-35(D0-D7)

```
申请152 Domain = 19 LD, >18, 必须从Rack LD0开始
LD 0 是Rack 0 LD0, 满足 → 分配 LD 0-18
Rack 0(LD 0-17) + Rack 1(LD 18)
结果: [0.0~18.7], Racks=[0, 1]
```

#### 示例12: >18 LD 从第二个Rack的LD0开始

可用: LD 6-35(D0-D7), LD 0-5 不可用

```
申请152 Domain = 19 LD, >18
候选: Rack 0 LD0 (LD 0) 不可用 → 跳过
候选: Rack 1 LD0 (LD 18) 可用 → 分配 LD 18-36
Rack 1(LD 18-35) + Rack 2(LD 36)
结果: [18.0~36.7], Racks=[1, 2]
```

#### 示例13: TPod 单Rack满足

可用Domain: LD 0-5(D0-D7), 全部在 Rack 0
TPod需求: [{"USB-HDSB", 2}]
可用TPod: {0: {"USB-HDSB": [0,1,2]}}

```
申请8 Domain, neededLDs=1
候选: LD 0(Rack 0, priority 0)
TPod检查: Rack 0满足USB-HDSB x2 → 分配 TPodId 0, 1
结果: [0.0~0.7], Racks=[0], TPods=[{0,0,"USB-HDSB"},{0,1,"USB-HDSB"}]
```

#### 示例14: TPod 第二个候选满足（规则5）

可用Domain: LD 0(D0-D7, Rack 0), LD 18(D0-D7, Rack 1)
TPod需求: [{"USB-HDSB", 2}]
可用TPod: {1: {"USB-HDSB": [0,1,2]}}  ← 只有 Rack 1 有 USB-HDSB

```
申请4 Domain, neededLDs=1
候选1: LD 0(Rack 0, priority 0) → Rack 0无USB-HDSB → 跳过
候选2: LD 18(Rack 1, priority 0) → Rack 1满足 → 成功
结果: [18.0~18.3], Racks=[1], TPods=[{1,0,"USB-HDSB"},{1,1,"USB-HDSB"}]
```

#### 示例15: 跨Rack Domain + 单Rack TPod

可用Domain: LD 0-35(D0-D7, Rack 0+1 全部)
TPod需求: [{"PCI", 1}]
可用TPod: {1: {"PCI": [3,4]}}  ← 只有 Rack 1 有 PCI

```
申请152 Domain = 19 LD, >6, 从LD 0开始
候选1: LD 0-18, Racks=[0,1]
TPod检查: Rack 0无PCI, Rack 1有PCI x1 → 分配 TPodId 3 from Rack 1
结果: [0.0~18.7], Racks=[0,1], TPods=[{1,3,"PCI"}]
（Domain跨Rack 0和1，TPod仅在Rack 1中分配，满足规则4）
```

#### 示例16: 所有候选都不满足TPod（失败）

可用Domain: LD 0-5(D0-D7, Rack 0)
TPod需求: [{"USB-HDSB", 2}]
可用TPod: {0: {"PCI": [0,1]}}  ← Rack 0只有PCI，没有USB-HDSB

```
申请8 Domain, neededLDs=1
候选: LD 0(Rack 0) → USB-HDSB不足 → 跳过
所有候选尝试完毕 → 分配失败
错误: cannot satisfy TPod requirements in any LD: LD 0 (rack 0): rack 0 cannot satisfy TPod requirement: USB-HDSB x2
```

### 错误场景

| 场景 | 错误信息 |
|-----|---------|
| 请求数量 <= 0 | `requested count must be positive, got {n}` |
| 可用设备列表为空 | `no available devices` |
| 单LD内无足够连续Domain | `cannot allocate {n} consecutive domains in any LD` |
| <=6 LD无法在单Cluster内分配 | `cannot allocate {n} consecutive LDs within a cluster` |
| >6 LD无法从Cluster LD0开始分配 | `cannot allocate {n} consecutive LDs starting from cluster LD0` |
| >18 LD无法从Rack LD0开始分配 | `cannot allocate {n} consecutive LDs starting from rack LD0` |
| Domain名称格式无效 | `invalid Domain name: {name}, expected format like 0.0` |
| Domain索引超出范围 | `invalid Domain index: {n} in {name}, must be 0-7` |
| 未知机型 | `unknown machine type: {s}` |
| TPod ExtType 为空 | `TPod requirement type must not be empty at index {i}` |
| TPod Number <= 0 | `TPod requirement number must be positive, got {n} for type "{type}"` |
| 有TPod需求但无可用TPod | `no available TPods provided` |
| 所有候选都无法满足TPod需求 | `cannot satisfy TPod requirements in any LD: {details}` 或 `cannot satisfy TPod requirements in any candidate: {details}` |
