// Package zebualloc 提供ZeBu设备自动分配功能。
//
// 支持zs3/zs4系统的HalfModule分配和zs5系统的SubModule分配。
// 用户只需声明所需HalfModule或SubModule的数量，该包即可从可用设备列表中选择合适的设备返回。
//
// 分配规则:
//   - 当Module数量大于等于4时，优先分配完整的Unit，剩余Module可在任意Unit内连续分配
//   - Unit之间的Module分配可以不连续
//   - Unit内的Module分配必须连续
//   - 当Module数量小于4时，必须从某个Unit的M0开始连续分配
package zebualloc
