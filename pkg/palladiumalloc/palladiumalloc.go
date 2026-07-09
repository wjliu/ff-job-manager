package palladiumalloc

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// MachineType 表示Palladium机型
type MachineType int

const (
	PZ1 MachineType = iota // Palladium Z1
	PZ2                    // Palladium Z2
)

const (
	domainsPerLD    = 8                              // 每个LD包含的Domain数量
	ldPerCluster    = 6                              // 每个Cluster包含的LD数量
	clustersPerRack = 3                              // 每个Rack包含的Cluster数量
	ldPerRack       = ldPerCluster * clustersPerRack // 18
)

// domainRe 匹配 Domain 格式: 0.0, 1.3, 12.7 等 (LD编号.Domain编号)
var domainRe = regexp.MustCompile(`^(\d+)\.(\d+)$`)

// ldDomains 表示一个LD内可用的Domain集合
type ldDomains struct {
	ldIndex int
	domains []int // 可用的Domain索引，已排序
}

// TPodRequirement 表示一个TPod外设需求
type TPodRequirement struct {
	ExtType string // 外设类型，如 "USB-HDSB", "PCI"
	Number  int    // 需要的该类型外设数量
}

// AllocatedTPod 表示一个已分配的TPod
type AllocatedTPod struct {
	RackId  int    // Rack编号
	TPodId  int    // TPod编号
	ExtType string // 外设类型
}

// String 返回MachineType的字符串表示
func (m MachineType) String() string {
	switch m {
	case PZ1:
		return "Palladium Z1"
	case PZ2:
		return "Palladium Z2"
	default:
		return fmt.Sprintf("unknown(%d)", m)
	}
}

// ParseMachineType 从字符串解析MachineType
func ParseMachineType(s string) (MachineType, error) {
	switch strings.ToLower(s) {
	case "pz1", "palladium z1":
		return PZ1, nil
	case "pz2", "palladium z2":
		return PZ2, nil
	default:
		return 0, fmt.Errorf("unknown machine type: %s", s)
	}
}

// Allocate 从可用设备列表中分配指定数量的Domain
//
// 参数:
//   - requestedCount: 需要的Domain数量
//   - available: 可用Domain名称列表，格式如 "0.0" (LD编号.Domain编号)
//   - machineType: 机型(PZ1/PZ2)
//
// 返回分配的Domain名称列表和涉及的Rack编号列表
func Allocate(requestedCount int, available []string, machineType MachineType) ([]string, []int, error) {
	domains, racks, _, err := AllocateWithTPod(requestedCount, available, machineType, nil, nil)
	return domains, racks, err
}

// AllocateWithTPod 从可用设备列表中分配指定数量的Domain，同时可选分配TPod外设
//
// 参数:
//   - requestedCount: 需要的Domain数量
//   - availableDomains: 可用Domain名称列表，格式如 "0.0" (LD编号.Domain编号)
//   - machineType: 机型(PZ1/PZ2)
//   - tpodReqs: TPod需求列表，nil或空数组表示不需要TPod分配
//   - availableTPods: 可用TPod信息，key为RackId，value为ExtType→TPodId数组的映射
//
// 返回分配的Domain名称列表、涉及的Rack编号列表和分配的TPod列表
func AllocateWithTPod(requestedCount int, availableDomains []string, machineType MachineType, tpodReqs []TPodRequirement, availableTPods map[int]map[string][]int) ([]string, []int, []AllocatedTPod, error) {
	if requestedCount <= 0 {
		return nil, nil, nil, fmt.Errorf("requested count must be positive, got %d", requestedCount)
	}
	if len(availableDomains) == 0 {
		return nil, nil, nil, fmt.Errorf("no available devices")
	}

	// 校验TPod需求参数
	if err := validateTPodReqs(tpodReqs); err != nil {
		return nil, nil, nil, err
	}
	hasTPod := len(tpodReqs) > 0
	if hasTPod && len(availableTPods) == 0 {
		return nil, nil, nil, fmt.Errorf("no available TPods provided")
	}

	// 解析可用列表，按LD分组
	ldMap, err := buildLDMap(availableDomains)
	if err != nil {
		return nil, nil, nil, err
	}

	// 计算所需LD数量和余数
	neededLDs := (requestedCount + domainsPerLD - 1) / domainsPerLD
	remainder := requestedCount % domainsPerLD
	if remainder == 0 {
		remainder = domainsPerLD // 整除时最后一个LD也需要全部8个
	}

	// 执行分配（流式：边遍历边检查Domain+TPod，满足即返回）
	var allocatedLDs []ldDomains
	var allocatedTPods []AllocatedTPod
	if neededLDs == 1 {
		// <=8 Domain: 在单个LD内分配
		ld, tpods, err := allocateWithinLD(requestedCount, ldMap, tpodReqs, availableTPods)
		if err != nil {
			return nil, nil, nil, err
		}
		allocatedLDs = []ldDomains{ld}
		allocatedTPods = tpods
	} else {
		// >8 Domain: 跨LD分配
		allocatedLDs, allocatedTPods, err = allocateAcrossLDs(neededLDs, remainder, ldMap, tpodReqs, availableTPods)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// 转换为Domain名称列表
	names := ldDomainsToNames(allocatedLDs)

	// 截取输出至实际请求数量（非倍数场景）
	if len(names) > requestedCount {
		names = names[:requestedCount]
	}

	// 计算涉及的Rack编号
	racks := computeRacks(allocatedLDs)

	return names, racks, allocatedTPods, nil
}

// buildLDMap 解析可用Domain列表，按LD分组返回可用的Domain
func buildLDMap(available []string) ([]ldDomains, error) {
	ldData := make(map[int]map[int]bool)

	for _, name := range available {
		m := domainRe.FindStringSubmatch(name)
		if m == nil {
			return nil, fmt.Errorf("invalid Domain name: %s, expected format like 0.0", name)
		}
		ldIdx := mustAtoi(m[1])
		domIdx := mustAtoi(m[2])
		if domIdx < 0 || domIdx >= domainsPerLD {
			return nil, fmt.Errorf("invalid Domain index: %d in %s, must be 0-%d", domIdx, name, domainsPerLD-1)
		}
		if ldData[ldIdx] == nil {
			ldData[ldIdx] = make(map[int]bool)
		}
		ldData[ldIdx][domIdx] = true
	}

	var result []ldDomains
	for ldIdx, doms := range ldData {
		var domainIndices []int
		for domIdx := range doms {
			domainIndices = append(domainIndices, domIdx)
		}
		sort.Ints(domainIndices)
		result = append(result, ldDomains{ldIndex: ldIdx, domains: domainIndices})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ldIndex < result[j].ldIndex
	})

	return result, nil
}

// allocateWithinLD 在单个LD内分配连续Domain，同时可选验证TPod
//
// 流式处理：在3个优先级循环中边遍历LD边检查TPod，满足即返回，不收集全部候选。
// 规则5：若当前LD的TPod不满足，继续尝试下一个LD，而非立即失败。
func allocateWithinLD(count int, ldMap []ldDomains, tpodReqs []TPodRequirement, availableTPods map[int]map[string][]int) (ldDomains, []AllocatedTPod, error) {
	hasTPod := len(tpodReqs) > 0
	seen := make(map[int]bool) // 高优先级已处理的LD不再在低优先级重复
	var tpodErrors []string

	// 优先级0: 全局搜索从D0开始的LD（碎片最小）
	for _, ld := range ldMap {
		if seen[ld.ldIndex] {
			continue
		}
		if len(ld.domains) < count {
			continue
		}
		if ld.domains[0] == 0 && isConsecutiveFrom(ld.domains, 0, count) {
			result := ldDomains{ldIndex: ld.ldIndex, domains: make([]int, count)}
			for i := 0; i < count; i++ {
				result.domains[i] = i
			}
			if hasTPod {
				tpods, err := allocateTPods([]int{ld.ldIndex / ldPerRack}, tpodReqs, availableTPods)
				if err != nil {
					tpodErrors = append(tpodErrors, fmt.Sprintf("LD %d (rack %d): %v", ld.ldIndex, ld.ldIndex/ldPerRack, err))
					seen[ld.ldIndex] = true
					continue // 规则5: TPod不满足，继续下一个候选
				}
				return result, tpods, nil
			}
			return result, nil, nil
		}
	}

	// 优先级1: 全局搜索末尾对齐D7的LD
	for _, ld := range ldMap {
		if seen[ld.ldIndex] {
			continue
		}
		if len(ld.domains) < count {
			continue
		}
		if ld.domains[len(ld.domains)-1] == domainsPerLD-1 {
			startDom := domainsPerLD - count
			if isConsecutiveFrom(ld.domains, startDom, count) {
				result := ldDomains{ldIndex: ld.ldIndex, domains: make([]int, count)}
				for i := 0; i < count; i++ {
					result.domains[i] = startDom + i
				}
				if hasTPod {
					tpods, err := allocateTPods([]int{ld.ldIndex / ldPerRack}, tpodReqs, availableTPods)
					if err != nil {
						tpodErrors = append(tpodErrors, fmt.Sprintf("LD %d (rack %d): %v", ld.ldIndex, ld.ldIndex/ldPerRack, err))
						seen[ld.ldIndex] = true
						continue // 规则5
					}
					return result, tpods, nil
				}
				return result, nil, nil
			}
		}
	}

	// 优先级2: 全局搜索任意连续段
	for _, ld := range ldMap {
		if seen[ld.ldIndex] {
			continue
		}
		if len(ld.domains) < count {
			continue
		}
		for i := 0; i <= len(ld.domains)-count; i++ {
			startDom := ld.domains[i]
			if isConsecutiveFrom(ld.domains, startDom, count) {
				result := ldDomains{ldIndex: ld.ldIndex, domains: make([]int, count)}
				for j := 0; j < count; j++ {
					result.domains[j] = startDom + j
				}
				if hasTPod {
					tpods, err := allocateTPods([]int{ld.ldIndex / ldPerRack}, tpodReqs, availableTPods)
					if err != nil {
						tpodErrors = append(tpodErrors, fmt.Sprintf("LD %d (rack %d): %v", ld.ldIndex, ld.ldIndex/ldPerRack, err))
						continue // 规则5
					}
					return result, tpods, nil
				}
				return result, nil, nil
			}
		}
	}

	// 所有候选尝试完毕
	if hasTPod && len(tpodErrors) > 0 {
		return ldDomains{}, nil, fmt.Errorf("cannot satisfy TPod requirements in any LD: %s", strings.Join(tpodErrors, "; "))
	}
	return ldDomains{}, nil, fmt.Errorf("cannot allocate %d consecutive domains in any LD", count)
}

// isConsecutiveFrom 检查从指定Domain开始是否有count个连续Domain可用
func isConsecutiveFrom(domains []int, startDom, count int) bool {
	idx := 0
	// 跳到startDom或之后的位置
	for idx < len(domains) && domains[idx] < startDom {
		idx++
	}
	for i := 0; i < count; i++ {
		expected := startDom + i
		if idx >= len(domains) || domains[idx] != expected {
			return false
		}
		idx++
	}
	return true
}

// isFullLD 检查LD是否完整（8个Domain全部可用）
func isFullLD(ld ldDomains) bool {
	return len(ld.domains) == domainsPerLD && ld.domains[0] == 0 && ld.domains[domainsPerLD-1] == domainsPerLD-1
}

// allocateAcrossLDs 跨LD分配Domain，同时可选验证TPod
//
// 流式处理：遍历排序后的起始LD，边尝试分配边检查TPod，满足即返回。
// 规则5：若当前候选的TPod不满足，继续尝试下一个起始LD。
func allocateAcrossLDs(neededLDs, remainder int, ldMap []ldDomains, tpodReqs []TPodRequirement, availableTPods map[int]map[string][]int) ([]ldDomains, []AllocatedTPod, error) {
	hasTPod := len(tpodReqs) > 0

	// 构建ldIndex到ldDomains的快速查找
	ldByIndex := make(map[int]ldDomains)
	for _, ld := range ldMap {
		ldByIndex[ld.ldIndex] = ld
	}

	// 收集所有可用的LD索引
	availableLDIndices := make(map[int]bool)
	for _, ld := range ldMap {
		availableLDIndices[ld.ldIndex] = true
	}

	// 确定起始LD的候选列表
	candidatesSet := make(map[int]bool)
	for _, ld := range ldMap {
		if neededLDs > ldPerRack {
			// >18 LD（跨Rack）: 起始必须是Rack的第一个Cluster的第一个LD
			if ld.ldIndex%ldPerRack == 0 {
				candidatesSet[ld.ldIndex] = true
			}
		} else if neededLDs > ldPerCluster {
			// >6 LD: 起始必须是Cluster的0号LD
			if ld.ldIndex%ldPerCluster == 0 {
				candidatesSet[ld.ldIndex] = true
			}
		} else {
			// <=6 LD: 可以从任何LD开始，但需在同一个Cluster内
			candidatesSet[ld.ldIndex] = true
		}
	}

	// 按优先级排序候选起始LD（碎片优先：Rack LD0 > Cluster LD0 > 其他）
	var sortedCandidates []int
	for c := range candidatesSet {
		sortedCandidates = append(sortedCandidates, c)
	}
	sort.Ints(sortedCandidates)
	sort.Slice(sortedCandidates, func(i, j int) bool {
		iAtRack := sortedCandidates[i]%ldPerRack == 0
		jAtRack := sortedCandidates[j]%ldPerRack == 0
		if iAtRack != jAtRack {
			return iAtRack
		}
		iAtCluster := sortedCandidates[i]%ldPerCluster == 0
		jAtCluster := sortedCandidates[j]%ldPerCluster == 0
		if iAtCluster != jAtCluster {
			return iAtCluster
		}
		return sortedCandidates[i] < sortedCandidates[j]
	})

	// 流式遍历：边尝试分配边检查TPod，满足即返回
	var tpodErrors []string
	for _, startLD := range sortedCandidates {
		lds, ok := tryAllocateLDs(startLD, neededLDs, remainder, ldByIndex, availableLDIndices)
		if !ok {
			continue
		}
		if hasTPod {
			racks := computeRacks(lds)
			tpods, err := allocateTPods(racks, tpodReqs, availableTPods)
			if err != nil {
				tpodErrors = append(tpodErrors, fmt.Sprintf("start LD %d racks %v: %v", startLD, racks, err))
				continue // 规则5: TPod不满足，继续下一个候选
			}
			return lds, tpods, nil
		}
		return lds, nil, nil
	}

	// 所有候选尝试完毕
	if hasTPod && len(tpodErrors) > 0 {
		return nil, nil, fmt.Errorf("cannot satisfy TPod requirements in any candidate: %s", strings.Join(tpodErrors, "; "))
	}
	if neededLDs > ldPerRack {
		return nil, nil, fmt.Errorf("cannot allocate %d consecutive LDs starting from rack LD0", neededLDs)
	}
	if neededLDs > ldPerCluster {
		return nil, nil, fmt.Errorf("cannot allocate %d consecutive LDs starting from cluster LD0", neededLDs)
	}
	return nil, nil, fmt.Errorf("cannot allocate %d consecutive LDs within a cluster", neededLDs)
}

// tryAllocateLDs 尝试从startLD开始分配neededLDs个连续LD
func tryAllocateLDs(startLD, neededLDs, remainder int, ldByIndex map[int]ldDomains, availableLDIndices map[int]bool) ([]ldDomains, bool) {
	clusterOfStart := startLD / ldPerCluster

	var result []ldDomains

	for i := 0; i < neededLDs; i++ {
		ldIdx := startLD + i

		// 检查LD是否在可用列表中
		if !availableLDIndices[ldIdx] {
			return nil, false
		}

		ld, exists := ldByIndex[ldIdx]
		if !exists {
			return nil, false
		}

		// Cluster约束: <=6 LD时必须在同一Cluster内
		if neededLDs <= ldPerCluster {
			if ld.ldIndex/ldPerCluster != clusterOfStart {
				return nil, false
			}
		}

		// >6 LD时起始必须是Cluster的0号LD
		if neededLDs > ldPerCluster && i == 0 && ld.ldIndex%ldPerCluster != 0 {
			return nil, false
		}

		// >18 LD（跨Rack）时起始必须是Rack的第一个Cluster的第一个LD
		if neededLDs > ldPerRack && i == 0 && ld.ldIndex%ldPerRack != 0 {
			return nil, false
		}

		if i < neededLDs-1 || remainder == domainsPerLD {
			// 完整LD: 必须全部8个Domain可用
			if !isFullLD(ld) {
				return nil, false
			}
			fullLD := ldDomains{ldIndex: ld.ldIndex, domains: make([]int, domainsPerLD)}
			for d := 0; d < domainsPerLD; d++ {
				fullLD.domains[d] = d
			}
			result = append(result, fullLD)
		} else {
			// 最后一个LD（部分）: 从D0开始连续remainder个Domain
			if !isConsecutiveFrom(ld.domains, 0, remainder) {
				return nil, false
			}
			partialLD := ldDomains{ldIndex: ld.ldIndex, domains: make([]int, remainder)}
			for d := 0; d < remainder; d++ {
				partialLD.domains[d] = d
			}
			result = append(result, partialLD)
		}
	}

	return result, true
}

// ldDomainsToNames 将分配的LD-Domain列表转换为Domain名称
func ldDomainsToNames(lds []ldDomains) []string {
	var names []string
	for _, ld := range lds {
		for _, dom := range ld.domains {
			names = append(names, fmt.Sprintf("%d.%d", ld.ldIndex, dom))
		}
	}
	return names
}

// computeRacks 计算分配涉及的Rack编号列表（去重排序）
func computeRacks(lds []ldDomains) []int {
	rackSet := make(map[int]bool)
	for _, ld := range lds {
		rackSet[ld.ldIndex/ldPerRack] = true
	}
	var racks []int
	for r := range rackSet {
		racks = append(racks, r)
	}
	sort.Ints(racks)
	return racks
}

// allocateTPods 从Rack列表中找一个能独立满足所有TPod需求的Rack（规则4：不跨Rack）
func allocateTPods(racks []int, tpodReqs []TPodRequirement, availableTPods map[int]map[string][]int) ([]AllocatedTPod, error) {
	for _, rackId := range racks {
		rackTPods, ok := availableTPods[rackId]
		if !ok {
			continue
		}
		tpods, err := allocateTPodsInRack(rackId, tpodReqs, rackTPods)
		if err != nil {
			continue
		}
		return tpods, nil
	}
	return nil, fmt.Errorf("no rack can satisfy all TPod requirements")
}

// allocateTPodsInRack 在单个Rack内分配TPod
func allocateTPodsInRack(rackId int, reqs []TPodRequirement, rackTPods map[string][]int) ([]AllocatedTPod, error) {
	var result []AllocatedTPod
	for _, req := range reqs {
		available, ok := rackTPods[req.ExtType]
		if !ok || len(available) < req.Number {
			return nil, fmt.Errorf("rack %d cannot satisfy TPod requirement: %s x%d", rackId, req.ExtType, req.Number)
		}
		for i := 0; i < req.Number; i++ {
			result = append(result, AllocatedTPod{
				RackId:  rackId,
				TPodId:  available[i],
				ExtType: req.ExtType,
			})
		}
	}
	return result, nil
}

// validateTPodReqs 校验TPod需求参数
func validateTPodReqs(tpodReqs []TPodRequirement) error {
	for i, req := range tpodReqs {
		if req.ExtType == "" {
			return fmt.Errorf("TPod requirement type must not be empty at index %d", i)
		}
		if req.Number <= 0 {
			return fmt.Errorf("TPod requirement number must be positive, got %d for type %q", req.Number, req.ExtType)
		}
	}
	return nil
}

// mustAtoi 将字符串转换为整数，调用者保证输入合法
func mustAtoi(s string) int {
	var n int
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
