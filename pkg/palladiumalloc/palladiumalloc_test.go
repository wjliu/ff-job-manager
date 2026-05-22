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
