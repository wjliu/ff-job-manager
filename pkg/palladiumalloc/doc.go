// Package palladiumalloc 提供Palladium设备自动分配功能。
//
// 支持Palladium Z1/Z2系统的Domain分配和可选的TPod外设分配。
// 用户只需声明所需Domain的数量（和可选的TPod外设需求），
// 该包即可从可用设备列表中选择合适的设备返回。
//
// 分配规则:
//   - Domain必须连续分配，<=8个Domain时不跨LD
//   - >8个Domain时跨LD，前面的LD必须全满，最后一个LD从D0开始连续
//   - LD必须连续分配，<=6个LD时不跨Cluster
//   - >6个LD时，起始LD必须是某Cluster的0号LD
//   - 尽量减少资源碎片（优先从D0和LD0起始）
//   - TPod分配可选，必须在Domain分配的Rack内满足，且不能跨Rack
//   - 若当前Domain候选的Rack不满足TPod，会继续尝试下一个候选
package palladiumalloc
