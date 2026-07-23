package zebualloc

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// SystemType 表示ZeBu系统类型
type SystemType int

const (
	ZS3 SystemType = iota // zs3系统，每个Module包含2个HalfModule
	ZS4                   // zs4系统，每个Module包含2个HalfModule
	ZS5                   // zs5系统，每个Module包含4个SubModule
)

// modulesPerUnit 每个Unit包含的Module数量
const modulesPerUnit = 4

// subModulesPerModule 每个Module包含的子模块数量
func subModulesPerModule(sysType SystemType) int {
	if sysType == ZS5 {
		return 4
	}
	return 2
}

// moduleUnit 表示一个Module所属的Unit索引和Module索引
type moduleUnit struct {
	unitIndex   int
	moduleIndex int
}

var (
	// 匹配 HalfModule 格式: U0.HM0, U1.HM3 等
	halfModuleRe = regexp.MustCompile(`^U(\d+)\.HM(\d+)$`)
	// 匹配 SubModule 格式: U0.M0.S0, U1.M2.S3 等
	subModuleRe = regexp.MustCompile(`^U(\d+)\.M(\d+)\.S(\d+)$`)
)

// parseHalfModule 解析HalfModule名称，返回其所属的Unit索引和Module索引
func parseHalfModule(name string) (moduleUnit, error) {
	m := halfModuleRe.FindStringSubmatch(name)
	if m == nil {
		return moduleUnit{}, fmt.Errorf("invalid HalfModule name: %s, expected format like U0.HM0", name)
	}
	unitIdx := mustAtoi(m[1])
	hmIdx := mustAtoi(m[2])
	// 每两个HalfModule组成一个Module: HM0,HM1->M0, HM2,HM3->M1
	modIdx := hmIdx / 2
	return moduleUnit{unitIndex: unitIdx, moduleIndex: modIdx}, nil
}

// parseSubModule 解析SubModule名称，返回其所属的Unit索引、Module索引和SubModule索引
func parseSubModule(name string) (moduleUnit, int, error) {
	m := subModuleRe.FindStringSubmatch(name)
	if m == nil {
		return moduleUnit{}, 0, fmt.Errorf("invalid SubModule name: %s, expected format like U0.M0.S0", name)
	}
	unitIdx := mustAtoi(m[1])
	modIdx := mustAtoi(m[2])
	subIdx := mustAtoi(m[3])
	return moduleUnit{unitIndex: unitIdx, moduleIndex: modIdx}, subIdx, nil
}

// unitModules 表示一个Unit内可用的Module集合
type unitModules struct {
	unitIndex int
	modules   []int // 可用的Module索引，已排序
}

// Allocate 从可用设备列表中分配指定数量的HalfModule或SubModule
//
// 参数:
//   - requestedCount: 需要的HalfModule(zs3/zs4)或SubModule(zs5)数量
//   - available: 可用设备名称列表
//   - sysType: 系统类型(ZS3/ZS4/ZS5)
//
// 返回分配的设备名称列表
func Allocate(requestedCount int, available []string, sysType SystemType) ([]string, error) {
	if requestedCount <= 0 {
		return nil, fmt.Errorf("requested count must be positive, got %d", requestedCount)
	}
	if len(available) == 0 {
		return nil, fmt.Errorf("no available devices")
	}

	subPerMod := subModulesPerModule(sysType)

	// 将请求的子模块数量转换为Module数量
	requestedModules := (requestedCount + subPerMod - 1) / subPerMod

	// 解析可用列表，构建Unit->Module的映射
	unitMap, err := buildUnitMap(available, sysType)
	if err != nil {
		return nil, err
	}

	// 检查可用Module总数是否足够
	availableModules := 0
	for _, um := range unitMap {
		availableModules += len(um.modules)
	}
	if availableModules < requestedModules {
		return nil, fmt.Errorf("not enough modules: requested %d, available %d", requestedModules, availableModules)
	}

	// 执行分配
	allocatedModules, err := allocateModules(requestedModules, unitMap)
	if err != nil {
		return nil, err
	}

	// 将Module分配结果转换回设备名称
	names := modulesToNames(allocatedModules, sysType)

	// 当请求数量不是subPerMod的倍数时，最后一个Module只有部分子模块被需要，截取实际请求数量
	if len(names) > requestedCount {
		names = names[:requestedCount]
	}

	return names, nil
}

// buildUnitMap 解析可用设备列表，按Unit分组返回可用的Module
func buildUnitMap(available []string, sysType SystemType) ([]unitModules, error) {
	// unitIndex -> moduleIndex -> subModuleIndices
	type modSet map[int]map[int]bool
	unitData := make(map[int]modSet)

	for _, name := range available {
		if sysType == ZS5 {
			mu, subIdx, err := parseSubModule(name)
			if err != nil {
				return nil, err
			}
			if unitData[mu.unitIndex] == nil {
				unitData[mu.unitIndex] = make(modSet)
			}
			if unitData[mu.unitIndex][mu.moduleIndex] == nil {
				unitData[mu.unitIndex][mu.moduleIndex] = make(map[int]bool)
			}
			unitData[mu.unitIndex][mu.moduleIndex][subIdx] = true
		} else {
			mu, err := parseHalfModule(name)
			if err != nil {
				return nil, err
			}
			if unitData[mu.unitIndex] == nil {
				unitData[mu.unitIndex] = make(modSet)
			}
			if unitData[mu.unitIndex][mu.moduleIndex] == nil {
				unitData[mu.unitIndex][mu.moduleIndex] = make(map[int]bool)
			}
			// HalfModule: 用hmIdx % 2作为子索引
			hmIdx := 0
			if m := halfModuleRe.FindStringSubmatch(name); m != nil {
				hmIdx = mustAtoi(m[2])
			}
			unitData[mu.unitIndex][mu.moduleIndex][hmIdx%2] = true
		}
	}

	subPerMod := subModulesPerModule(sysType)
	var result []unitModules
	for unitIdx, mods := range unitData {
		var moduleIndices []int
		for modIdx, subs := range mods {
			// 只有当Module内所有子模块都可用时，该Module才算可用
			if len(subs) == subPerMod {
				moduleIndices = append(moduleIndices, modIdx)
			}
		}
		sort.Ints(moduleIndices)
		if len(moduleIndices) > 0 {
			result = append(result, unitModules{unitIndex: unitIdx, modules: moduleIndices})
		}
	}

	// 按Unit索引排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].unitIndex < result[j].unitIndex
	})

	return result, nil
}

// allocateModules 执行Module维度的分配
func allocateModules(requestedModules int, unitMap []unitModules) ([]moduleUnit, error) {
	var allocated []moduleUnit

	if requestedModules >= modulesPerUnit {
		// 优先分配完整Unit
		completeUnits := requestedModules / modulesPerUnit
		remaining := requestedModules % modulesPerUnit

		for _, um := range unitMap {
			if completeUnits <= 0 {
				break
			}
			if len(um.modules) >= modulesPerUnit {
				// 检查Module是否从M0开始连续
				if isConsecutiveFromZero(um.modules, modulesPerUnit) {
					for i := 0; i < modulesPerUnit; i++ {
						allocated = append(allocated, moduleUnit{unitIndex: um.unitIndex, moduleIndex: i})
					}
					completeUnits--
				}
			}
		}

		if completeUnits > 0 {
			return nil, fmt.Errorf("not enough complete units: need %d more", completeUnits)
		}

		if remaining > 0 {
			// 剩余Module从任意可用Unit内分配，需连续（支持M3→M0回转）
			remainingMods, err := allocateRemaining(remaining, unitMap, allocated)
			if err != nil {
				return nil, err
			}
			allocated = append(allocated, remainingMods...)
		}
	} else {
		// Module数 < 4: 在某个Unit内连续分配（支持M3→M0回转）
		remainingMods, err := allocateRemaining(requestedModules, unitMap, nil)
		if err != nil {
			return nil, err
		}
		allocated = append(allocated, remainingMods...)
	}

	return allocated, nil
}

// isConsecutiveFromZero 检查Module列表是否从0开始连续包含n个Module
func isConsecutiveFromZero(modules []int, n int) bool {
	if len(modules) < n {
		return false
	}
	for i := 0; i < n; i++ {
		if modules[i] != i {
			return false
		}
	}
	return true
}

// isAlreadyAllocated 检查某个Module是否已被分配
func isAlreadyAllocated(allocated []moduleUnit, unitIdx, modIdx int) bool {
	for _, a := range allocated {
		if a.unitIndex == unitIdx && a.moduleIndex == modIdx {
			return true
		}
	}
	return false
}

// allocateRemaining 分配剩余的Module，在Unit内连续分配（支持M3→M0回转）。
//
// 为减少资源碎片化，当存在多个Unit都能满足分配需求时，优先选择可用Module数量
// 最少的Unit，以尽量保留完整Unit供后续大数量请求使用。若数量相同则按Unit索引
// 从小到大选择以保证确定性。
func allocateRemaining(count int, unitMap []unitModules, allocated []moduleUnit) ([]moduleUnit, error) {
	// 收集所有有效的分配方案
	type candidate struct {
		modules    []moduleUnit
		availCount int // 该Unit的可用Module总数
		unitIndex  int
	}
	var candidates []candidate

	for _, um := range unitMap {
		// 尝试从每个可用Module作为起点，寻找连续count个Module（支持回转）
		for startIdx, startMod := range um.modules {
			if isAlreadyAllocated(allocated, um.unitIndex, startMod) {
				continue
			}
			result := tryConsecutiveWrap(um.modules, startIdx, count, allocated, um.unitIndex)
			if result != nil {
				candidates = append(candidates, candidate{
					modules:    result,
					availCount: len(um.modules),
					unitIndex:  um.unitIndex,
				})
				break // 同一个Unit找到第一个有效方案即可
			}
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("cannot allocate %d consecutive modules in any unit", count)
	}

	// 按碎片化优先级排序:
	//   - 优先选可用Module数少的Unit（减少碎片，保留完整Unit）
	//   - 同数量时按Unit索引排序（确定性）
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].availCount != candidates[j].availCount {
			return candidates[i].availCount < candidates[j].availCount
		}
		return candidates[i].unitIndex < candidates[j].unitIndex
	})

	return candidates[0].modules, nil
}

// tryConsecutiveWrap 尝试从modules[startIdx]开始分配count个连续Module（支持M3→M0回转）
func tryConsecutiveWrap(modules []int, startIdx, count int, allocated []moduleUnit, unitIndex int) []moduleUnit {
	var result []moduleUnit
	modIdx := modules[startIdx]

	for i := 0; i < count; i++ {
		// 回转: M3之后回到M0
		wrappedModIdx := modIdx % modulesPerUnit
		if isAlreadyAllocated(allocated, unitIndex, wrappedModIdx) {
			return nil
		}
		if !containsModule(modules, wrappedModIdx) {
			return nil
		}
		result = append(result, moduleUnit{unitIndex: unitIndex, moduleIndex: wrappedModIdx})
		modIdx = wrappedModIdx + 1
	}

	return result
}

// containsModule 检查Module列表中是否包含指定索引
func containsModule(modules []int, modIdx int) bool {
	for _, m := range modules {
		if m == modIdx {
			return true
		}
	}
	return false
}

// modulesToNames 将分配的Module列表转换为设备名称
func modulesToNames(modules []moduleUnit, sysType SystemType) []string {
	subPerMod := subModulesPerModule(sysType)
	var names []string

	for _, mu := range modules {
		if sysType == ZS5 {
			for s := 0; s < subPerMod; s++ {
				names = append(names, fmt.Sprintf("U%d.M%d.S%d", mu.unitIndex, mu.moduleIndex, s))
			}
		} else {
			for s := 0; s < subPerMod; s++ {
				names = append(names, fmt.Sprintf("U%d.HM%d", mu.unitIndex, mu.moduleIndex*2+s))
			}
		}
	}

	return names
}

// mustAtoi 将字符串转换为整数，调用者保证输入合法
func mustAtoi(s string) int {
	var n int
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}

// FormatEnvVar 将分配的设备数组格式化为环境变量注入格式。
//
// 根据环境变量注入格式规范，将细粒度的 HalfModule 或 SubModule 尽可能合并为 Module 粒度，
// 完整 Unit 仅注入 Ux.M0，部分 Unit 注入第一个被分配 Module，并以逗号拼接。
// 输出顺序：完整 Unit 的 M0 在前，零散 Unit 的 Module 在后。
//
// 参数:
//   - allocated: Allocate 函数返回的设备名称列表
//   - sysType: 系统类型 (ZS3/ZS4/ZS5)
//
// 返回格式化后的字符串，例如 "U1.M0,U0.M2"
func FormatEnvVar(allocated []string, sysType SystemType) string {
	if len(allocated) == 0 {
		return ""
	}

	modOrder, moduleMap := parseAllocatedModules(allocated, sysType)
	unitOrder, unitMap := groupModulesByUnit(modOrder, moduleMap)
	return buildEnvVarOutput(unitOrder, unitMap, sysType)
}

// allocatedModInfo 记录每个 Module 的分配信息
type allocatedModInfo struct {
	unitIndex   int
	moduleIndex int
	subModules  map[int]bool // 有哪些子模块被分配
	complete    bool         // 该 Module 的所有子模块是否都被分配
}

// parseAllocatedModules 解析分配的设备名，按 Module 粒度归并并标记完整性
func parseAllocatedModules(allocated []string, sysType SystemType) ([]string, map[string]*allocatedModInfo) {
	subPerMod := subModulesPerModule(sysType)
	moduleMap := make(map[string]*allocatedModInfo) // key: "U{u}.M{m}"
	var modOrder []string                           // 保持 Module 首次出现顺序

	for _, name := range allocated {
		unitIdx, modIdx, subIdx := parseDeviceName(name, sysType)
		if unitIdx < 0 {
			continue // 跳过非法名称
		}

		key := fmt.Sprintf("U%d.M%d", unitIdx, modIdx)
		if _, exists := moduleMap[key]; !exists {
			moduleMap[key] = &allocatedModInfo{
				unitIndex:   unitIdx,
				moduleIndex: modIdx,
				subModules:  make(map[int]bool),
			}
			modOrder = append(modOrder, key)
		}
		moduleMap[key].subModules[subIdx] = true
	}

	// 标记每个 Module 是否完整（所有子模块都被分配）
	for _, mi := range moduleMap {
		mi.complete = len(mi.subModules) == subPerMod
	}

	return modOrder, moduleMap
}

// parseDeviceName 解析单个设备名称，返回 unitIndex, moduleIndex, subIndex。
// 解析失败时 unitIndex 返回 -1。
func parseDeviceName(name string, sysType SystemType) (unitIdx, modIdx, subIdx int) {
	if sysType == ZS5 {
		mu, sIdx, err := parseSubModule(name)
		if err != nil {
			return -1, 0, 0
		}
		return mu.unitIndex, mu.moduleIndex, sIdx
	}

	mu, err := parseHalfModule(name)
	if err != nil {
		return -1, 0, 0
	}
	// 从 HM 名称解析子模块索引: HM0→0, HM1→1, HM2→0, HM3→1, ...
	if m := halfModuleRe.FindStringSubmatch(name); m != nil {
		return mu.unitIndex, mu.moduleIndex, mustAtoi(m[2]) % 2
	}
	return mu.unitIndex, mu.moduleIndex, 0
}

// unitAllocData 记录一个 Unit 内的分配信息
type unitAllocData struct {
	unitIndex int
	modules   []*allocatedModInfo // 该 Unit 内的 Module，按首次出现顺序
}

// groupModulesByUnit 将 Module 按 Unit 分组，保持首次出现顺序
func groupModulesByUnit(modOrder []string, moduleMap map[string]*allocatedModInfo) ([]int, map[int]*unitAllocData) {
	unitMap := make(map[int]*unitAllocData)
	var unitOrder []int

	for _, key := range modOrder {
		mi := moduleMap[key]
		if _, exists := unitMap[mi.unitIndex]; !exists {
			unitMap[mi.unitIndex] = &unitAllocData{
				unitIndex: mi.unitIndex,
			}
			unitOrder = append(unitOrder, mi.unitIndex)
		}
		unitMap[mi.unitIndex].modules = append(unitMap[mi.unitIndex].modules, mi)
	}

	return unitOrder, unitMap
}

// buildEnvVarOutput 构建最终的环境变量注入字符串
func buildEnvVarOutput(unitOrder []int, unitMap map[int]*unitAllocData, sysType SystemType) string {
	var completeUnitSegs []string
	var partialUnitSegs []string
	var incompleteSegs []string

	for _, unitIdx := range unitOrder {
		ud := unitMap[unitIdx]
		completeCount := countCompleteModules(ud.modules)

		if completeCount == modulesPerUnit {
			// 完整 Unit：仅注入 Ux.M0
			completeUnitSegs = append(completeUnitSegs, fmt.Sprintf("U%d.M0", ud.unitIndex))
		} else {
			// 部分 Unit：注入第一个被分配的完整 Module
			if first := firstAllocatedModule(ud.modules); first != nil {
				partialUnitSegs = append(partialUnitSegs,
					fmt.Sprintf("U%d.M%d", first.unitIndex, first.moduleIndex))
			}
			// 不完整的 Module 保持独立设备名格式
			incompleteSegs = appendIncompleteSegs(incompleteSegs, ud.modules, sysType)
		}
	}

	// 组合输出：完整 Unit 在前，部分 Unit 次之，独立设备名最后
	var allSegs []string
	allSegs = append(allSegs, completeUnitSegs...)
	allSegs = append(allSegs, partialUnitSegs...)
	allSegs = append(allSegs, incompleteSegs...)

	return strings.Join(allSegs, ",")
}

// countCompleteModules 统计完整 Module 的数量
func countCompleteModules(modules []*allocatedModInfo) int {
	count := 0
	for _, mi := range modules {
		if mi.complete {
			count++
		}
	}
	return count
}

// firstAllocatedModule 根据分配语义返回部分 Unit 中第一个被分配的完整 Module。
//
// 分配算法保证每个 Unit 内的 Module 是一个连续的环状段。通过找到环上的
// 缺口（缺失的 Module），缺口之后第一个存在的 Module 即为分配起点。
// 例如 [M3, M0]：缺口在 M1，M3 是起点；[M0, M1]：缺口在 M2，M0 是起点。
func firstAllocatedModule(modules []*allocatedModInfo) *allocatedModInfo {
	// 收集完整 Module 的索引
	var indices []int
	for _, mi := range modules {
		if mi.complete {
			indices = append(indices, mi.moduleIndex)
		}
	}
	if len(indices) == 0 {
		return nil
	}

	sort.Ints(indices)

	// 在环 [M0,M1,M2,M3] 上找缺口，缺口后的第一个 Module 即为分配起点
	startIdx := findConsecutiveStart(indices)
	for _, mi := range modules {
		if mi.complete && mi.moduleIndex == startIdx {
			return mi
		}
	}
	return nil
}

// findConsecutiveStart 在环 M0→M1→M2→M3→M0 上找到第一个缺口，
// 返回缺口后第一个存在的 Module 索引（即连续段的起点）。
func findConsecutiveStart(indices []int) int {
	for i := 0; i < modulesPerUnit; i++ {
		if !containsInt(indices, i) {
			// i 是缺口，向后（环状）找第一个存在的 Module
			for j := 1; j <= modulesPerUnit; j++ {
				candidate := (i + j) % modulesPerUnit
				if containsInt(indices, candidate) {
					return candidate
				}
			}
		}
	}
	// 无缺口（完整 Unit，4 个 Module 全有），不应从此路径进入
	return 0
}

// containsInt 检查切片是否包含指定整数
func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// appendIncompleteSegs 将不完整 Module 的子模块按排序追加到 segments 中
func appendIncompleteSegs(segs []string, modules []*allocatedModInfo, sysType SystemType) []string {
	for _, mi := range modules {
		if mi.complete {
			continue
		}
		// 按子模块索引排序输出，保证确定性
		subIdxList := make([]int, 0, len(mi.subModules))
		for s := range mi.subModules {
			subIdxList = append(subIdxList, s)
		}
		sort.Ints(subIdxList)
		for _, subIdx := range subIdxList {
			segs = append(segs, formatSubDevice(mi.unitIndex, mi.moduleIndex, subIdx, sysType))
		}
	}
	return segs
}

// formatSubDevice 根据系统类型格式化单个子模块设备名
func formatSubDevice(unitIdx, modIdx, subIdx int, sysType SystemType) string {
	if sysType == ZS5 {
		return fmt.Sprintf("U%d.M%d.S%d", unitIdx, modIdx, subIdx)
	}
	return fmt.Sprintf("U%d.HM%d", unitIdx, modIdx*2+subIdx)
}

// String 返回SystemType的字符串表示
func (s SystemType) String() string {
	switch s {
	case ZS3:
		return "zs3"
	case ZS4:
		return "zs4"
	case ZS5:
		return "zs5"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// ParseSystemType 从字符串解析SystemType
func ParseSystemType(s string) (SystemType, error) {
	switch strings.ToLower(s) {
	case "zs3":
		return ZS3, nil
	case "zs4":
		return ZS4, nil
	case "zs5":
		return ZS5, nil
	default:
		return 0, fmt.Errorf("unknown system type: %s", s)
	}
}
