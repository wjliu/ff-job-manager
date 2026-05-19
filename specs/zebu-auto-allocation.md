# 自动分配设备

## 需求描述

该功能旨在为作业自动选择ZeBu设备，用户只需要在提交作业时声明需要HalfModule (zs3或zs4) 或者SubModule (zs5) 的数量，该工具便可以在可用的HalfModule和SubModule列表中选择出适合的设备返回。

## 背景知识

ZeBu是Synopsys的一款用于做芯片验证的硬件仿真系统，其不同型号中的概念如下：
- zs3或zs4：ZeBu系统由多个Unit组成，每个Unit包含4个Module，每个Module包含两个HalfModule。Unit可以使用U0、U1、U2这样的命名表示，从U0开始命名；Module可以使用U0.M0、U0.M1、U0.M2这样的命名表示，从U0.M0开始命名；HalfModule可以使用U0.HM0、U0.HM1、U0.HM2这样的命名表示，从U0.HM0开始命名。其中U0.HM0和U0.HM1对应组合成U0.M0，依此类推。

- zs5：ZeBu系统由多个Unit组成，每个Unit包含4个Module，每个Module包含四个SubModule。Unit可以使用U0、U1、U2这样的命名表示，从U0开始命名；Module可以使用U0.M0、U0.M1、U0.M2这样的命名表示，从U0.M0开始命名；SubModule可以使用U0.M0.S0、U0.M0.S1、U0.M0.S2这样的命名表>示，从U0.M0.S0开始命名。其中U0.M0.S0、U0.M0.S1、U0.M0.S2、U0.M0.S3对应组合成U0.M0，依此类推。

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
1. 当HalfModule或者SubModule的申请数量转换成Module数量后大于4时，应保证前面N个4的倍数的数量部分都是分配完整的Unit，后面不足4个的Module可以在任意的Unit内的任一个Module开始分配，不限制必须在M0的位置开始分配
2. 在1的基础上分配时，应保证Unit之间的Module分配可以不连续，即分配完U0后，不用必须在U1上分配，可以跳到U2或者其他Unit上分配
3. 在1的基础上分配时，应保证Unit之内的Module分配必须是连续的，即从U0.M0分配完后，必须连续分配U0.M1，不能直接跳到U0.M2
4. 当HalfModule或者SubModule的申请数量转换成Module数量后小于4时，应保证Module的分配是从某个U的M0开始分配且保证连续



