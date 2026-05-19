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
	return modulesToNames(allocatedModules, sysType), nil
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
			// 剩余Module从任意可用Unit内分配，需连续但不要求从M0开始
			remainingMods, err := allocateRemaining(remaining, unitMap, allocated, false)
			if err != nil {
				return nil, err
			}
			allocated = append(allocated, remainingMods...)
		}
	} else {
		// Module数 < 4: 必须从某个Unit的M0开始连续分配
		remainingMods, err := allocateRemaining(requestedModules, unitMap, nil, true)
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

// allocateRemaining 分配剩余的Module
// fromM0: 是否必须从M0开始
func allocateRemaining(count int, unitMap []unitModules, allocated []moduleUnit, fromM0 bool) ([]moduleUnit, error) {
	for _, um := range unitMap {
		// 在此Unit中寻找连续的可用Module
		startIdx := 0
		if fromM0 {
			// 必须从M0开始
			if len(um.modules) == 0 || um.modules[0] != 0 {
				continue
			}
			startIdx = 0
		} else {
			// 可以从任意位置开始
			startIdx = -1
			for i := 0; i <= len(um.modules)-count; i++ {
				// 检查从um.modules[i]开始是否有count个连续且未被分配的Module
				if checkConsecutive(um.modules, i, count, allocated, um.unitIndex) {
					startIdx = i
					break
				}
			}
			if startIdx == -1 {
				continue
			}
		}

		// 尝试从startIdx开始分配count个连续Module
		var result []moduleUnit
		modIdx := um.modules[startIdx]
		allocatedCount := 0

		for allocatedCount < count {
			if isAlreadyAllocated(allocated, um.unitIndex, modIdx) {
				break
			}
			if !containsModule(um.modules, modIdx) {
				break
			}
			result = append(result, moduleUnit{unitIndex: um.unitIndex, moduleIndex: modIdx})
			allocatedCount++
			modIdx++
		}

		if allocatedCount == count {
			return result, nil
		}
	}

	if fromM0 {
		return nil, fmt.Errorf("cannot allocate %d consecutive modules from M0", count)
	}
	return nil, fmt.Errorf("cannot allocate %d consecutive modules in any unit", count)
}

// checkConsecutive 检查从modules[startPos]开始是否有count个连续且未被分配的Module
func checkConsecutive(modules []int, startPos, count int, allocated []moduleUnit, unitIndex int) bool {
	expectedModIdx := modules[startPos]
	j := startPos
	for i := 0; i < count; i++ {
		if j >= len(modules) {
			return false
		}
		if modules[j] != expectedModIdx {
			return false
		}
		if isAlreadyAllocated(allocated, unitIndex, expectedModIdx) {
			return false
		}
		expectedModIdx++
		j++
	}
	return true
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
