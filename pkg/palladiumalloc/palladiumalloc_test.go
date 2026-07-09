package palladiumalloc

import (
	"fmt"
	"reflect"
	"testing"
)

// 生成指定LD数量的完整可用Domain列表
func makeFullAvailable(ldCount int) []string {
	var avail []string
	for ld := 0; ld < ldCount; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	return avail
}

func TestAllocate_SingleDomain(t *testing.T) {
	avail := makeFullAvailable(1)
	result, racks, err := Allocate(1, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"0.0"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
	expectedRacks := []int{0}
	if !reflect.DeepEqual(racks, expectedRacks) {
		t.Errorf("expected racks %v, got %v", expectedRacks, racks)
	}
}

func TestAllocate_4DomainsInOneLD(t *testing.T) {
	avail := makeFullAvailable(1)
	result, _, err := Allocate(4, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 碎片优先: 从D0开始
	expected := []string{"0.0", "0.1", "0.2", "0.3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_8DomainsOneFullLD(t *testing.T) {
	avail := makeFullAvailable(2)
	result, _, err := Allocate(8, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"0.0", "0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_9DomainsCrossLD(t *testing.T) {
	avail := makeFullAvailable(2)
	result, _, err := Allocate(9, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// LD0全部 + LD1.D0
	expected := []string{
		"0.0", "0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7",
		"1.0",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_16DomainsTwoFullLDs(t *testing.T) {
	avail := makeFullAvailable(3)
	result, _, err := Allocate(16, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{
		"0.0", "0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7",
		"1.0", "1.1", "1.2", "1.3", "1.4", "1.5", "1.6", "1.7",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_48DomainsOneCluster(t *testing.T) {
	avail := makeFullAvailable(6) // 1个完整Cluster
	result, racks, err := Allocate(48, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 48 {
		t.Errorf("expected 48 Domains, got %d", len(result))
	}
	expectedRacks := []int{0}
	if !reflect.DeepEqual(racks, expectedRacks) {
		t.Errorf("expected racks %v, got %v", expectedRacks, racks)
	}
}

func TestAllocate_49DomainsCrossCluster(t *testing.T) {
	// 49 Domain = 7 LD, >6, 必须从Cluster的0号LD开始
	avail := makeFullAvailable(18) // 3个Rack完整
	result, racks, err := Allocate(49, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 49 {
		t.Errorf("expected 49 Domains, got %d", len(result))
	}
	// 涉及Rack 0 (LD 0-5) 和 Rack 0 (LD 6, Cluster 1 也在 Rack 0)
	// Cluster 0: LD 0-5, Cluster 1: LD 6-11, 都在 Rack 0
	expectedRacks := []int{0}
	if !reflect.DeepEqual(racks, expectedRacks) {
		t.Errorf("expected racks %v, got %v", expectedRacks, racks)
	}
}

func TestAllocate_CrossRack(t *testing.T) {
	// 18 LD = 1 Rack, 分配144 Domain
	avail := makeFullAvailable(36) // 2 Racks
	result, racks, err := Allocate(144, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 144 {
		t.Errorf("expected 144 Domains, got %d", len(result))
	}
	// 144 Domain = 18 LD = 1 Rack (LD 0-17)
	expectedRacks := []int{0}
	if !reflect.DeepEqual(racks, expectedRacks) {
		t.Errorf("expected racks %v, got %v", expectedRacks, racks)
	}
}

func TestAllocate_LessThan6LDsInCluster(t *testing.T) {
	// 3 LDs within one Cluster
	avail := makeFullAvailable(6)
	result, _, err := Allocate(24, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 24 {
		t.Errorf("expected 24 Domains, got %d", len(result))
	}
	// 应从LD 0开始 (碎片优先)
	if result[0] != "0.0" {
		t.Errorf("expected start from LD 0, got %s", result[0])
	}
}

func TestAllocate_MoreThan6LDsMustStartFromClusterLD0(t *testing.T) {
	// 7 LDs, >6, 必须从Cluster的0号LD开始
	avail := makeFullAvailable(18)
	result, _, err := Allocate(56, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 56 {
		t.Errorf("expected 56 Domains, got %d", len(result))
	}
	// 必须从LD 0 (Cluster 0的0号LD) 开始
	if result[0] != "0.0" {
		t.Errorf("expected start from LD 0, got %s", result[0])
	}
}

func TestAllocate_MoreThan6LDsCannotStartFromNonClusterLD0(t *testing.T) {
	// LD 0-5 不可用, 只有 LD 1-17 可用
	var avail []string
	for ld := 1; ld < 18; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	// 7 LDs 需要, 但 LD 0 不可用, LD 6 (Cluster 1 LD0) 可用
	_, _, err := Allocate(56, avail, PZ1)
	if err != nil {
		t.Fatalf("should succeed starting from LD 6: %v", err)
	}
}

func TestAllocate_FragmentationPrefersD0(t *testing.T) {
	// LD 0 有 D4-D7 可用, LD 1 有 D0-D7 可用
	avail := []string{
		"0.4", "0.5", "0.6", "0.7",
		"1.0", "1.1", "1.2", "1.3", "1.4", "1.5", "1.6", "1.7",
	}
	// 申请4个Domain, 优先选从D0开始的LD
	result, _, err := Allocate(4, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"1.0", "1.1", "1.2", "1.3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v (D0 start), got %v", expected, result)
	}
}

func TestAllocate_FragmentationPrefersEdgeD7(t *testing.T) {
	// LD 0 有 D4-D7 可用, 没有 D0 起始的LD
	avail := []string{"0.4", "0.5", "0.6", "0.7"}
	result, _, err := Allocate(4, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 末尾对齐D7
	expected := []string{"0.4", "0.5", "0.6", "0.7"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_NonMultipleCount(t *testing.T) {
	avail := makeFullAvailable(2)

	// 申请3个Domain
	result, _, err := Allocate(3, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 Domains, got %d: %v", len(result), result)
	}
	expected3 := []string{"0.0", "0.1", "0.2"}
	if !reflect.DeepEqual(result, expected3) {
		t.Errorf("expected %v, got %v", expected3, result)
	}

	// 申请5个Domain
	result, _, err = Allocate(5, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("expected 5 Domains, got %d: %v", len(result), result)
	}

	// 申请10个Domain
	result, _, err = Allocate(10, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 10 {
		t.Errorf("expected 10 Domains, got %d: %v", len(result), result)
	}
}

func TestAllocate_ZeroCount(t *testing.T) {
	avail := makeFullAvailable(1)
	_, _, err := Allocate(0, avail, PZ1)
	if err == nil {
		t.Fatal("expected error for zero count")
	}
}

func TestAllocate_EmptyAvailable(t *testing.T) {
	_, _, err := Allocate(1, []string{}, PZ1)
	if err == nil {
		t.Fatal("expected error for empty available list")
	}
}

func TestAllocate_InvalidDomainName(t *testing.T) {
	avail := []string{"0.0", "INVALID"}
	_, _, err := Allocate(1, avail, PZ1)
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
}

func TestAllocate_InvalidDomainIndex(t *testing.T) {
	avail := []string{"0.8"} // Domain index 8 超出范围
	_, _, err := Allocate(1, avail, PZ1)
	if err == nil {
		t.Fatal("expected error for invalid domain index")
	}
}

func TestAllocate_InsufficientDomains(t *testing.T) {
	avail := []string{"0.0", "0.1"} // 只有2个Domain
	_, _, err := Allocate(3, avail, PZ1)
	if err == nil {
		t.Fatal("expected error for insufficient domains")
	}
}

func TestAllocate_InsufficientFullLDs(t *testing.T) {
	// LD 0 完整, LD 1 只有 D0-D3
	avail := []string{
		"0.0", "0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7",
		"1.0", "1.1", "1.2", "1.3",
	}
	// 需要2个完整LD, 但LD 1不完整
	_, _, err := Allocate(16, avail, PZ1)
	if err == nil {
		t.Fatal("expected error for insufficient full LDs")
	}
}

func TestAllocate_PartialLastLDMustStartFromD0(t *testing.T) {
	// LD 0 完整, LD 1 只有 D2-D7 (D0, D1 不可用)
	avail := []string{
		"0.0", "0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7",
		"1.2", "1.3", "1.4", "1.5", "1.6", "1.7",
	}
	// 需要10 Domain = 1完整LD + 2 Domain (D0-D1)
	// LD 1 的 D0-D1 不可用, 无法满足
	_, _, err := Allocate(10, avail, PZ1)
	if err == nil {
		t.Fatal("expected error: partial LD must start from D0")
	}
}

func TestAllocate_LDsNotConsecutive(t *testing.T) {
	// LD 0 和 LD 2 可用, LD 1 不可用
	avail := []string{
		"0.0", "0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7",
		"2.0", "2.1", "2.2", "2.3", "2.4", "2.5", "2.6", "2.7",
	}
	// 需要16 Domain = 2完整LD, 但LD不连续
	_, _, err := Allocate(16, avail, PZ1)
	if err == nil {
		t.Fatal("expected error: LDs not consecutive")
	}
}

func TestAllocate_MoreThan6LDsNotStartFromClusterLD0(t *testing.T) {
	// LD 1-12 可用, LD 0 不可用
	var avail []string
	for ld := 1; ld <= 12; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	// 7 LDs, >6, 需从Cluster LD0开始
	// LD 1 不是Cluster LD0 (Cluster 0 LD0 是 LD 0), 但 LD 6 是 Cluster 1 LD0
	_, _, err := Allocate(56, avail, PZ1)
	if err != nil {
		t.Fatalf("should succeed from LD 6 (Cluster 1 LD0): %v", err)
	}
}

func TestAllocate_RackComputation(t *testing.T) {
	// 分配跨Rack: LD 17 (Rack 0) 和 LD 18 (Rack 1)
	// Rack 0: LD 0-17, Rack 1: LD 18-35
	avail := makeFullAvailable(36) // 2 Racks
	// 分配19 LD (LD 0-18, 跨越Rack 0和Rack 1)
	// >6 LD, 必须从Cluster LD0开始, LD 0 满足
	_, racks, err := Allocate(19*8, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedRacks := []int{0, 1}
	if !reflect.DeepEqual(racks, expectedRacks) {
		t.Errorf("expected racks %v, got %v", expectedRacks, racks)
	}
}

func TestParseMachineType(t *testing.T) {
	tests := []struct {
		input string
		want  MachineType
	}{
		{"pz1", PZ1},
		{"PZ1", PZ1},
		{"Palladium Z1", PZ1},
		{"palladium z1", PZ1},
		{"pz2", PZ2},
		{"PZ2", PZ2},
		{"Palladium Z2", PZ2},
	}
	for _, tt := range tests {
		got, err := ParseMachineType(tt.input)
		if err != nil {
			t.Errorf("ParseMachineType(%q): unexpected error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("ParseMachineType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}

	_, err := ParseMachineType("unknown")
	if err == nil {
		t.Error("expected error for unknown machine type")
	}
}

func TestMachineTypeString(t *testing.T) {
	if PZ1.String() != "Palladium Z1" {
		t.Errorf("PZ1.String() = %q, want %q", PZ1.String(), "Palladium Z1")
	}
	if PZ2.String() != "Palladium Z2" {
		t.Errorf("PZ2.String() = %q, want %q", PZ2.String(), "Palladium Z2")
	}
}

// ========== TPod 分配测试 ==========

func TestAllocateWithTPod_SingleRackSatisfies(t *testing.T) {
	// 单Rack内同时满足Domain和TPod
	avail := makeFullAvailable(6) // LD 0-5, 全部在Rack 0
	tpodReqs := []TPodRequirement{
		{ExtType: "USB-HDSB", Number: 2},
	}
	availableTPods := map[int]map[string][]int{
		0: {"USB-HDSB": {0, 1, 2}},
	}

	domains, racks, tpods, err := AllocateWithTPod(8, avail, PZ1, tpodReqs, availableTPods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 8 {
		t.Errorf("expected 8 domains, got %d", len(domains))
	}
	if !reflect.DeepEqual(racks, []int{0}) {
		t.Errorf("expected racks [0], got %v", racks)
	}
	expectedTPods := []AllocatedTPod{
		{RackId: 0, TPodId: 0, ExtType: "USB-HDSB"},
		{RackId: 0, TPodId: 1, ExtType: "USB-HDSB"},
	}
	if !reflect.DeepEqual(tpods, expectedTPods) {
		t.Errorf("expected tpods %v, got %v", expectedTPods, tpods)
	}
}

func TestAllocateWithTPod_CrossRackDomainSingleRackTPod(t *testing.T) {
	// Domain跨Rack [0,1]，TPod只在Rack 1满足
	avail := makeFullAvailable(36) // LD 0-35, Rack 0: LD 0-17, Rack 1: LD 18-35
	tpodReqs := []TPodRequirement{
		{ExtType: "PCI", Number: 1},
	}
	availableTPods := map[int]map[string][]int{
		1: {"PCI": {3, 4}}, // 只在Rack 1有PCI
	}

	// 申请 19 LD (152 Domain) > 6, 从LD 0开始, 跨Rack 0和Rack 1
	// Rack 0无PCI，但Rack 1有PCI → 应成功
	domains, racks, tpods, err := AllocateWithTPod(152, avail, PZ1, tpodReqs, availableTPods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 152 {
		t.Errorf("expected 152 domains, got %d", len(domains))
	}
	if !reflect.DeepEqual(racks, []int{0, 1}) {
		t.Errorf("expected racks [0, 1], got %v", racks)
	}
	// TPod应从Rack 1分配
	if len(tpods) != 1 || tpods[0].RackId != 1 || tpods[0].ExtType != "PCI" {
		t.Errorf("expected TPod from Rack 1, got %v", tpods)
	}
}

func TestAllocateWithTPod_SecondCandidateSatisfiesTPod(t *testing.T) {
	// 规则5关键测试: 第一个Domain候选的Rack不满足TPod，第二个候选满足
	// LD 0 (Rack 0) D0-D7可用, LD 18 (Rack 1) D0-D7可用
	avail := []string{
		"0.0", "0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7",
		"18.0", "18.1", "18.2", "18.3", "18.4", "18.5", "18.6", "18.7",
	}
	tpodReqs := []TPodRequirement{
		{ExtType: "USB-HDSB", Number: 2},
	}
	// Rack 0 没有USB-HDSB, Rack 1 有
	availableTPods := map[int]map[string][]int{
		1: {"USB-HDSB": {0, 1, 2}},
	}

	// 申请4 Domain, neededLDs=1
	// 候选1: LD 0 (Rack 0, priority 0) → TPod不满足
	// 候选2: LD 18 (Rack 1, priority 0) → TPod满足 → 应成功
	domains, racks, tpods, err := AllocateWithTPod(4, avail, PZ1, tpodReqs, availableTPods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 应选择LD 18 (Rack 1)
	if domains[0] != "18.0" {
		t.Errorf("expected allocation from LD 18 (Rack 1), got %v", domains)
	}
	if !reflect.DeepEqual(racks, []int{1}) {
		t.Errorf("expected racks [1], got %v", racks)
	}
	if len(tpods) != 2 {
		t.Errorf("expected 2 TPods, got %d", len(tpods))
	}
}

func TestAllocateWithTPod_NoCandidateSatisfiesTPod(t *testing.T) {
	// 所有Domain候选的Rack都不满足TPod
	avail := makeFullAvailable(6) // LD 0-5, Rack 0
	tpodReqs := []TPodRequirement{
		{ExtType: "USB-HDSB", Number: 2},
	}
	// Rack 0 没有USB-HDSB类型
	availableTPods := map[int]map[string][]int{
		0: {"PCI": {0, 1}}, // 只有PCI，没有USB-HDSB
	}

	_, _, _, err := AllocateWithTPod(8, avail, PZ1, tpodReqs, availableTPods)
	if err == nil {
		t.Fatal("expected error: no candidate satisfies TPod")
	}
}

func TestAllocateWithTPod_MultipleTPodTypes(t *testing.T) {
	// 多种TPod类型同时需求
	avail := makeFullAvailable(6) // Rack 0
	tpodReqs := []TPodRequirement{
		{ExtType: "USB-HDSB", Number: 2},
		{ExtType: "PCI", Number: 1},
	}
	availableTPods := map[int]map[string][]int{
		0: {
			"USB-HDSB": {0, 1, 2},
			"PCI":      {3, 4},
		},
	}

	_, _, tpods, err := AllocateWithTPod(8, avail, PZ1, tpodReqs, availableTPods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tpods) != 3 {
		t.Errorf("expected 3 TPods (2 USB-HDSB + 1 PCI), got %d", len(tpods))
	}

	// 验证类型和数量
	usbCount := 0
	pciCount := 0
	for _, tp := range tpods {
		switch tp.ExtType {
		case "USB-HDSB":
			usbCount++
		case "PCI":
			pciCount++
		}
	}
	if usbCount != 2 {
		t.Errorf("expected 2 USB-HDSB, got %d", usbCount)
	}
	if pciCount != 1 {
		t.Errorf("expected 1 PCI, got %d", pciCount)
	}
}

func TestAllocateWithTPod_NilTPodReqs(t *testing.T) {
	// TPod需求为nil时行为与Allocate一致
	avail := makeFullAvailable(2)
	result, racks, tpods, err := AllocateWithTPod(4, avail, PZ1, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"0.0", "0.1", "0.2", "0.3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
	if !reflect.DeepEqual(racks, []int{0}) {
		t.Errorf("expected racks [0], got %v", racks)
	}
	if tpods != nil {
		t.Errorf("expected nil tpods, got %v", tpods)
	}
}

func TestAllocateWithTPod_EmptyTPodReqs(t *testing.T) {
	// TPod需求为空数组时行为与Allocate一致
	avail := makeFullAvailable(2)
	result, racks, tpods, err := AllocateWithTPod(4, avail, PZ1, []TPodRequirement{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"0.0", "0.1", "0.2", "0.3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
	if !reflect.DeepEqual(racks, []int{0}) {
		t.Errorf("expected racks [0], got %v", racks)
	}
	if tpods != nil {
		t.Errorf("expected nil tpods, got %v", tpods)
	}
}

func TestAllocateWithTPod_InvalidTPodNumber(t *testing.T) {
	avail := makeFullAvailable(1)
	tpodReqs := []TPodRequirement{
		{ExtType: "USB", Number: 0}, // Number <= 0
	}
	_, _, _, err := AllocateWithTPod(1, avail, PZ1, tpodReqs, nil)
	if err == nil {
		t.Fatal("expected error for TPod Number <= 0")
	}
}

func TestAllocateWithTPod_EmptyExtType(t *testing.T) {
	avail := makeFullAvailable(1)
	tpodReqs := []TPodRequirement{
		{ExtType: "", Number: 1}, // ExtType为空
	}
	_, _, _, err := AllocateWithTPod(1, avail, PZ1, tpodReqs, nil)
	if err == nil {
		t.Fatal("expected error for empty ExtType")
	}
}

func TestAllocateWithTPod_ReqsButNoAvailable(t *testing.T) {
	// 有TPod需求但availableTPods为nil
	avail := makeFullAvailable(1)
	tpodReqs := []TPodRequirement{
		{ExtType: "USB", Number: 1},
	}
	_, _, _, err := AllocateWithTPod(1, avail, PZ1, tpodReqs, nil)
	if err == nil {
		t.Fatal("expected error: no available TPods provided")
	}
}

func TestAllocateWithTPod_ReqsButEmptyAvailable(t *testing.T) {
	// 有TPod需求但availableTPods为空map
	avail := makeFullAvailable(1)
	tpodReqs := []TPodRequirement{
		{ExtType: "USB", Number: 1},
	}
	_, _, _, err := AllocateWithTPod(1, avail, PZ1, tpodReqs, map[int]map[string][]int{})
	if err == nil {
		t.Fatal("expected error: no available TPods provided")
	}
}

func TestAllocateWithTPod_InsufficientTPodQuantity(t *testing.T) {
	// TPod数量不足
	avail := makeFullAvailable(6) // Rack 0
	tpodReqs := []TPodRequirement{
		{ExtType: "USB-HDSB", Number: 5}, // 需要5个
	}
	availableTPods := map[int]map[string][]int{
		0: {"USB-HDSB": {0, 1, 2}}, // 只有3个
	}

	_, _, _, err := AllocateWithTPod(8, avail, PZ1, tpodReqs, availableTPods)
	if err == nil {
		t.Fatal("expected error: insufficient TPod quantity")
	}
}

func TestAllocateWithTPod_CrossLDSecondCandidateSatisfiesTPod(t *testing.T) {
	// 规则5 + 跨LD: 第一个跨LD候选不满足TPod，第二个候选满足
	// LD 0-5 (Rack 0, Cluster 0) 和 LD 6-11 (Rack 0, Cluster 1) 都完整
	// 但 LD 0-5 和 LD 6-11 都在 Rack 0（最多3个Cluster/Rack）
	// 需要制造跨Rack场景:
	// Rack 0: LD 0-17, Rack 1: LD 18-35
	// 候选1: LD 0-6 (Rack 0, 7个LD, >6) → Rack 0无TPod → 跳过
	// 候选2: LD 18-24 (Rack 1, 7个LD, >6) → Rack 1有TPod → 成功

	var avail []string
	// Rack 0 LD 0-17 全部可用, Rack 1 LD 18-35 全部可用
	for ld := 0; ld < 36; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}

	tpodReqs := []TPodRequirement{
		{ExtType: "USB-HDSB", Number: 2},
	}
	// Rack 0没有USB-HDSB, Rack 1有
	availableTPods := map[int]map[string][]int{
		1: {"USB-HDSB": {0, 1, 2}},
	}

	// >6 LD (7 LD), 排序优先Rack LD0
	// 候选: LD 0(Rack LD0) → LD 0-6, Rack 0, TPod失败 → 跳过
	// 候选: LD 18(Rack LD0) → LD 18-24, Rack 1, TPod满足 → 成功
	domains, racks, tpods, err := AllocateWithTPod(56, avail, PZ1, tpodReqs, availableTPods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// LD 18-24 全部在Rack 1内
	if len(domains) != 56 {
		t.Errorf("expected 56 domains, got %d", len(domains))
	}
	if !reflect.DeepEqual(racks, []int{1}) {
		t.Errorf("expected racks [1], got %v", racks)
	}
	// TPod从Rack 1分配
	if len(tpods) != 2 || tpods[0].RackId != 1 {
		t.Errorf("expected 2 TPods from Rack 1, got %v", tpods)
	}
}

func TestAllocate_MoreThan18LDsMustStartFromRackLD0(t *testing.T) {
	// 19 LDs (>18, 跨Rack), 必须从Rack LD0开始
	// LD 0-35全部可用
	avail := makeFullAvailable(36)
	// 申请152 Domain = 19 LD
	result, racks, err := Allocate(152, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 152 {
		t.Errorf("expected 152 domains, got %d", len(result))
	}
	// 必须从LD 0开始（Rack 0 LD0）
	if result[0] != "0.0" {
		t.Errorf("expected start from LD 0 (Rack LD0), got %s", result[0])
	}
	// 跨Rack: Rack 0 (LD 0-17) + Rack 1 (LD 18)
	expectedRacks := []int{0, 1}
	if !reflect.DeepEqual(racks, expectedRacks) {
		t.Errorf("expected racks %v, got %v", expectedRacks, racks)
	}
}

func TestAllocate_MoreThan18LDsCannotStartFromNonRackLD0(t *testing.T) {
	// LD 0-5 不可用 (Rack 0 LD0所在的Cluster 0不可用)
	// LD 6-17 可用 (Rack 0 剩余), LD 18-36 可用 (Rack 1 LD0 + Rack 2 LD0)
	var avail []string
	for ld := 6; ld < 37; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	// 申请152 Domain = 19 LD (>18)
	// LD 0(Rack 0 LD0)不可用 → 跳过
	// LD 18(Rack 1 LD0)可用, LD 18-36 19个LD都可用 → 应从LD 18开始
	result, racks, err := Allocate(152, avail, PZ1)
	if err != nil {
		t.Fatalf("should succeed from LD 18 (Rack 1 LD0): %v", err)
	}
	if len(result) != 152 {
		t.Errorf("expected 152 domains, got %d", len(result))
	}
	if result[0] != "18.0" {
		t.Errorf("expected start from LD 18 (Rack 1 LD0), got %s", result[0])
	}
	// LD 18-36: Rack 1(18-35) + Rack 2(36)
	expectedRacks := []int{1, 2}
	if !reflect.DeepEqual(racks, expectedRacks) {
		t.Errorf("expected racks [1, 2], got %v", racks)
	}
}

func TestAllocate_MoreThan18LDsNoRackLD0Available(t *testing.T) {
	// 所有Rack LD0 (0, 18, 36...) 都不可用
	// LD 1-17 (Rack 0, 不含LD 0), LD 19-35 (Rack 1, 不含LD 18)
	var avail []string
	for ld := 1; ld < 18; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	for ld := 19; ld < 36; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	// 申请152 Domain = 19 LD (>18)
	// 没有Rack LD0可用 → 应失败
	_, _, err := Allocate(152, avail, PZ1)
	if err == nil {
		t.Fatal("expected error: must start from Rack LD0 for >18 LDs")
	}
}

func TestAllocate_MoreThan18LDsFromSecondRackLD0(t *testing.T) {
	// LD 18-54可用（Rack 1 LD0=18, Rack 2 LD0=36都可用）
	var avail []string
	for ld := 18; ld < 54; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	// 申请152 Domain = 19 LD (>18)
	// 优先级: LD 18(Rack 1 LD0) > LD 36(Rack 2 LD0)
	// 应从LD 18开始
	result, racks, err := Allocate(152, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 152 {
		t.Errorf("expected 152 domains, got %d", len(result))
	}
	if result[0] != "18.0" {
		t.Errorf("expected start from LD 18, got %s", result[0])
	}
	// LD 18-36: Rack 1(18-35) + Rack 2(36)
	expectedRacks := []int{1, 2}
	if !reflect.DeepEqual(racks, expectedRacks) {
		t.Errorf("expected racks [1, 2], got %v", racks)
	}
}
