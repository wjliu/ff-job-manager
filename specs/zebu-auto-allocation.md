# 自动分配设备

## 需求描述

该功能旨在为作业自动选择ZeBu设备，用户只需要在提交作业时声明需要HalfModule (zs3或zs4) 或者SubModule (zs5) 的数量，该工具便可以在可用的HalfModule和SubModule列表中选择出适合的设备返回。

## 背景知识

ZeBu是Synopsys的一款用于做芯片验证的硬件仿真系统，其不同型号中的概念如下：
- zs3或zs4：ZeBu系统由多个Unit组成，每个Unit包含4个Module，每个Module包含两个HalfModule。Unit可以使用U0、U1、U2这样的命名表示，从U0开始命名；Module可以使用U0.M0、U0.M1、U0.M2这样的命名表示，从U0.M0开始命名；HalfModule可以使用U0.HM0、U0.HM1、U0.HM2这样的命名表示，从U0.HM0开始命名。其中U0.HM0和U0.HM1对应组合成U0.M0，依此类推。

- zs5：ZeBu系统由多个Unit组成，每个Unit包含4个Module，每个Module包含四个SubModule。Unit可以使用U0、U1、U2这样的命名表示，从U0开始命名；Module可以使用U0.M0、U0.M1、U0.M2这样的命名表示，从U0.M0开始命名；SubModule可以使用U0.M0.S0、U0.M0.S1、U0.M0.S2、U0.M0.S3这样的命名表示，从U0.M0.S0开始命名。其中U0.M0.S0、U0.M0.S1、U0.M0.S2、U0.M0.S3对应组合成U0.M0，依此类推。

## 输入

zs3或zs4：
- HalfModule的数量
- 可用的HalfModule数组，格式如：[U0.HM0, U0.HM1,U0.HM2,U0.HM3]

zs5：
- SubModule的数量
- 可用的SubModule数组，格式如：[U0.M0.S0, U0.M0.S1,U0.M0.S2,U0.M0.S3]

## 输出

zs3或zs4：
- 分配的HalfModule数组，格式如：[U0.HM0, U0.HM1,U0.HM2,U0.HM3]

zs5：
- 分配的SubModule数组，格式如：[U0.M0.S0, U0.M0.S1,U0.M0.S2,U0.M0.S3]

## 分配规则

在实际分配时应遵守如下规则：
1. 当HalfModule或者SubModule的申请数量转换成Module数量后大于等于4时，应保证前面N个4的倍数的数量部分都是分配完整的Unit，后面不足4个的Module可以在任意的Unit内的任一个Module开始分配，不限制必须在M0的位置开始分配
2. 在1的基础上分配时，应保证Unit之间的Module分配可以不连续，即分配完U0后，不用必须在U1上分配，可以跳到U2或者其他Unit上分配
3. 在1的基础上分配时，应保证Unit之内的Module分配必须是连续的，即从U0.M0分配完后，必须连续分配U0.M1，不能直接跳到U0.M2
4. 当HalfModule或者SubModule的申请数量转换成Module数量后小于4时，应保证Module的分配是在某个Unit内连续的，不能出现跳过Module的分配
5. 关于Unit内部的Module连续定义，除了常规的序号连续场景，同时分配到M0和M3这种收尾回转的也算是连续，因此分配时需要注意

## 分配机制

### 整体流程

```
输入: requestedCount, available[], sysType
  │
  ├─ 1. 参数校验（数量 > 0, 列表非空）
  │
  ├─ 2. 子模块数量 → Module数量转换
  │     requestedModules = ceil(requestedCount / subModulesPerModule)
  │
  ├─ 3. 构建Unit-Module映射（buildUnitMap）
  │     解析可用设备 → 按Unit分组 → 过滤不完整Module
  │
  ├─ 4. 检查可用Module总数是否 >= 请求Module数
  │
  ├─ 5. 执行Module维度分配（allocateModules）
  │     ├─ requestedModules >= 4: 优先分配完整Unit → 分配剩余Module
  │     └─ requestedModules < 4:  在某个Unit内连续分配（支持M3→M0回转）
  │
  ├─ 6. 将分配结果转换为设备名称输出
  │
  └─ 7. 截取输出至实际请求数量（非倍数场景）
```

### 步骤详解

#### 步骤1: 子模块到Module的转换

根据系统类型确定每个Module包含的子模块数量：

| 系统类型 | 子模块类型 | 每个Module包含的子模块数 |
|---------|-----------|----------------------|
| zs3     | HalfModule | 2                    |
| zs4     | HalfModule | 2                    |
| zs5     | SubModule  | 4                    |

转换公式：`requestedModules = ceil(requestedCount / subModulesPerModule)`

例如：zs3系统申请6个HalfModule → 6/2 = 3个Module；申请5个HalfModule → ceil(5/2) = 3个Module。

#### 步骤2: 构建Unit-Module映射

解析可用设备列表，按Unit分组，并过滤掉不完整的Module：

1. 遍历可用设备列表，解析每个设备名称，提取Unit索引和Module索引
2. 按Unit索引分组，在每个Unit内按Module索引收集子模块
3. **完整性过滤**：只有当一个Module内的所有子模块都可用时，该Module才算可用。例如zs3中U0.HM0和U0.HM1都存在，U0.M0才算可用；若只有U0.HM0，则U0.M0不可用
4. 按Unit索引排序，每个Unit内的Module索引也排序

#### 步骤3: Module维度分配

根据请求的Module数量，分配过程分为两种情况：

**情况A: requestedModules >= 4**

```
1. 计算需要分配的完整Unit数: completeUnits = requestedModules / 4
2. 计算剩余Module数:         remaining = requestedModules % 4

3. 分配完整Unit（按Unit索引从小到大）:
   遍历Unit映射，寻找满足以下条件的Unit:
   - 该Unit有4个可用Module
   - 4个Module从M0开始连续（即M0, M1, M2, M3均可用）
   找到后，将该Unit的全部4个Module标记为已分配
   重复直到分配完 completeUnits 个完整Unit

4. 若 remaining > 0，分配剩余Module:
   在任意Unit内寻找连续 remaining 个未分配的Module
   无需从M0开始，只要在Unit内连续即可（支持M3→M0回转）
```

**情况B: requestedModules < 4**

```
遍历Unit映射，寻找满足以下条件的Unit:
- 该Unit内有连续 requestedModules 个可用Module
- 连续分配的起点不限，无需从M0开始
- 支持M3→M0回转（即M3和M0视为相邻连续）
找到后，分配这些Module
```

#### 步骤4: 结果转换

将分配的Module列表按系统类型转换回设备名称：

- zs3/zs4: Module `U{u}.M{m}` → `U{u}.HM{m*2}`, `U{u}.HM{m*2+1}`
- zs5: Module `U{u}.M{m}` → `U{u}.M{m}.S0`, `U{u}.M{m}.S1`, `U{u}.M{m}.S2`, `U{u}.M{m}.S3`

#### 步骤5: 截取输出

当请求数量不是 `subModulesPerModule` 的倍数时，步骤4中 `modulesToNames` 会输出最后一个Module的全部子模块，导致返回数量超过请求。因此在输出前需截取至实际请求的数量 `requestedCount`。

例如：zs3申请3个HalfModule → 分配2个Module → modulesToNames输出4个HalfModule → 截取前3个返回。

### 分配示例

#### 示例1: zs3 申请2个HalfModule（1个Module，< 4规则）

可用设备: U0(U0.HM0~U0.HM7), U1(U1.HM0~U1.HM7)

```
转换: 2 HalfModule → 1 Module
分配: requestedModules(1) < 4, 在Unit内连续分配
结果: U0.M0 → [U0.HM0, U0.HM1]
```

#### 示例2: zs3 申请8个HalfModule（4个Module，= 4规则）

可用设备: U0(U0.HM0~U0.HM7), U1(U1.HM0~U1.HM7)

```
转换: 8 HalfModule → 4 Module
分配: 1个完整Unit
结果: U0.M0~U0.M3 → [U0.HM0, U0.HM1, U0.HM2, U0.HM3, U0.HM4, U0.HM5, U0.HM6, U0.HM7]
```

#### 示例3: zs3 申请10个HalfModule（5个Module，>= 4规则）

可用设备: U0(U0.HM0~U0.HM7), U1(U1.HM0~U1.HM7)

```
转换: 10 HalfModule → 5 Module
分配: 1个完整Unit(U0) + 1个剩余Module(U1.M0)
结果: [U0.HM0~U0.HM7, U1.HM0, U1.HM1]
```

#### 示例4: zs5 申请16个SubModule（4个Module，= 4规则）

可用设备: U0(U0.M0.S0~U0.M3.S3), U1(U1.M0.S0~U1.M3.S3)

```
转换: 16 SubModule → 4 Module
分配: 1个完整Unit(U0)
结果: U0.M0~U0.M3 → [U0.M0.S0~U0.M0.S3, U0.M1.S0~U0.M1.S3, U0.M2.S0~U0.M2.S3, U0.M3.S0~U0.M3.S3]
```

#### 示例5: Unit跳选（非连续Unit）

可用设备: U1(U1.HM0~U1.HM7), U3(U3.HM0~U3.HM7)，U0和U2不可用

```
申请: 8 HalfModule → 4 Module → 1个完整Unit
分配: U1满足完整Unit条件，跳过U0直接选U1
结果: [U1.HM0~U1.HM7]
```

#### 示例6: 剩余Module从非M0位置分配

可用设备: U0(U0.HM0~U0.HM7), U1(仅U1.HM2, U1.HM3，即U1.M1)

```
申请: 10 HalfModule → 5 Module
分配: 1个完整Unit(U0) + 1个剩余Module
      U1中M0不可用，但M1可用，剩余1个Module可从M1开始分配
结果: [U0.HM0~U0.HM7, U1.HM2, U1.HM3]
```

#### 示例7: 不完整Module被过滤

可用设备: U0.HM0(缺少U0.HM1，U0.M0不完整), U0.HM2, U0.HM3(U0.M1完整), U1.HM0, U1.HM1(U1.M0完整)

```
申请: 2 HalfModule → 1 Module, < 4规则
      U0.M0不完整，不可用；U0.M1完整可用（<4不要求M0起），优先选U0.M1
结果: [U0.HM2, U0.HM3]
```

#### 示例8: < 4规则不要求从M0开始

可用设备: U0.HM2, U0.HM3(U0.M1)

```
申请: 2 HalfModule → 1 Module, < 4规则
      U0.M1不从M0开始，但<4规则不要求从M0开始，可选
结果: [U0.HM2, U0.HM3]
```

#### 示例9: M3→M0回转连续分配

可用设备: U0.HM6, U0.HM7(U0.M3), U0.HM0, U0.HM1(U0.M0)

```
申请: 4 HalfModule → 2 Module, < 4规则
      M3和M0视为相邻连续，从M3开始回转到M0
结果: [U0.HM6, U0.HM7, U0.HM0, U0.HM1]
```

#### 示例10: 非倍数请求数量（截取输出）

可用设备: U0(U0.HM0~U0.HM7)

```
申请: 3 HalfModule（不是2的倍数）
转换: ceil(3/2) = 2 Module
分配: 2个Module (U0.M0, U0.M1)
转换名称: [U0.HM0, U0.HM1, U0.HM2, U0.HM3] (4个)
截取: 取前3个 → [U0.HM0, U0.HM1, U0.HM2]
```

#### 示例11: zs5非倍数请求数量

可用设备: U0(U0.M0.S0~U0.M3.S3)

```
申请: 6 SubModule（不是4的倍数）
转换: ceil(6/4) = 2 Module
分配: 2个Module (U0.M0, U0.M1)
转换名称: [U0.M0.S0~U0.M0.S3, U0.M1.S0~U0.M1.S3] (8个)
截取: 取前6个 → [U0.M0.S0, U0.M0.S1, U0.M0.S2, U0.M0.S3, U0.M1.S0, U0.M1.S1]
```

### 错误场景

| 场景 | 错误信息 |
|-----|---------|
| 请求数量 <= 0 | `requested count must be positive, got {n}` |
| 可用设备列表为空 | `no available devices` |
| 可用Module总数不足 | `not enough modules: requested {n}, available {m}` |
| 完整Unit不足 | `not enough complete units: need {n} more` |
| < 4规则下无法在任意Unit内找到连续Module（含回转） | `cannot allocate {n} consecutive modules in any unit` |
| >= 4规则下剩余Module无法在任意Unit内连续分配 | `cannot allocate {n} consecutive modules in any unit` |
| 设备名称格式无效 | `invalid HalfModule name: {name}` 或 `invalid SubModule name: {name}` |

## 环境变量注入格式

在基于分配机制成功分配设备后，会将分配的设备信息以环境变量形式注入，但是在注入时构造出的环境变量的格式需要遵守以下规则：
1. 应尽量以Module的粒度进行格式化，即细粒度的HalfModule或者SubModule应尽可能合并成Module的格式后注入。比如将U0.HM0,U0.HM1合并成U0.M0。
2. 某个Unit被完整分配时，仅需要注入第一个Module的编号即可，比如U0.M0、U0.M1、U0.M2、U0.M3都被分配了，仅注入U0.M0即可。
3. 某个Unit被部分分配时，仅需要注入第一个被分配的Module的编号即可，比如U0.M2、U0.M3被分配了，仅注入U0.M2即可。需要注意：如果分配的是U0.M3和U0.M0，则需要注入U0.M3而不是U0.M0，即首尾连接的场景，需要以分配的第一个Module为准，而不是M0。
4. 如果分配了多个Unit的Module时，应将完整分配的Unit的M0注入放在前面，将零散的Unit的Module注入放在后面，比如分配的是：U0.M2、U0.M3、U1.M0、U1.M1、U1.M2、U1.M3，则应注入U1.M0,U0.M2。
5. 注入的格式中如果包含多个Ux.My的注入，则应使用逗号拼接。
6. 需要注意zs3、zs4、zs5中HalfModule和SubModule的格式差异，使得在合并成Module时能够正确处理。