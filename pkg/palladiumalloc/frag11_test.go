package palladiumalloc

import (
	"fmt"
	"testing"
)

// 用户精确场景：
// LD 0: 仅 D0 被占用 (D1-D7 可用)
// LD 1: D0-D3 被占用 (D4-D7 可用)
// LD 2-17: 完整可用
// LD 29: 完整可用 (Rack 1, 唯一可用)
// LD 36-53: 完整可用 (Rack 2)

func TestAllocate_UserSecondAllocation_PartialLDs(t *testing.T) {
	var avail []string

	// Rack 1: LD 29 (仅此1个LD在Rack 1可用, 8 domains)
	for d := 0; d < domainsPerLD; d++ {
		avail = append(avail, fmt.Sprintf("%d.%d", 29, d))
	}

	// Rack 0: LD 0 (D1-D7, 仅D0被占)
	for d := 1; d < domainsPerLD; d++ {
		avail = append(avail, fmt.Sprintf("%d.%d", 0, d))
	}
	// Rack 0: LD 1 (D4-D7, D0-D3被占)
	for d := 4; d < domainsPerLD; d++ {
		avail = append(avail, fmt.Sprintf("%d.%d", 1, d))
	}
	// Rack 0: LD 2-17 (完整, 16个LD)
	for ld := 2; ld <= 17; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}

	// Rack 2: LD 36-53 (完整, 18个LD)
	for ld := 36; ld <= 53; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}

	// 按Domain数统计各Rack:
	// Rack 0: 7+4+128 = 139 domains
	// Rack 1: 8 domains
	// Rack 2: 144 domains
	// 碎片排序: Rack 1(8) → 无Cluster LD0候选
	//           Rack 0(139) < Rack 2(144) → Rack 0优先

	result, racks, err := Allocate(88, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("result[0]=%s, racks=%v", result[0], racks)

	// 候选: LD 0(Rack LD0, Cluster LD0) → try[0,11] LD0非完整 → 失败
	// 候选: LD 6(Rack 0, Cluster LD0) → try[6,11] LD6-16完整 → 成功!
	if result[0] != "6.0" {
		t.Errorf("expected start from LD 6 (Rack 0), got %s", result[0])
	}
}

// 同样场景但LD 0也完整（仅LD 1残缺）
func TestAllocate_UserSecondAllocation_OnlyLD1Partial(t *testing.T) {
	var avail []string

	// Rack 1: LD 29
	for d := 0; d < domainsPerLD; d++ {
		avail = append(avail, fmt.Sprintf("%d.%d", 29, d))
	}

	// Rack 0: LD 0 完整, LD 1 残缺(D4-D7), LD 2-17 完整
	for d := 0; d < domainsPerLD; d++ {
		avail = append(avail, fmt.Sprintf("%d.%d", 0, d)) // LD 0 完整
	}
	for d := 4; d < domainsPerLD; d++ {
		avail = append(avail, fmt.Sprintf("%d.%d", 1, d)) // LD 1 残缺
	}
	for ld := 2; ld <= 17; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}

	// Rack 2: LD 36-53
	for ld := 36; ld <= 53; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}

	// Rack 0: 8+4+128 = 140 domains < Rack 2: 144 → Rack 0 优先
	result, racks, err := Allocate(88, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("result[0]=%s, racks=%v", result[0], racks)

	// LD 0 完整(Rack LD0) → try[0,11] LD1不完整 → 失败
	// LD 6 → try[6,11] → 成功
	if result[0] != "6.0" {
		t.Errorf("expected start from LD 6 (Rack 0), got %s", result[0])
	}
}

// 13 LD 场景：<=18 LD 时不能跨 Rack
// Rack 0: LD 0-1 残缺, LD 2-17 完整
// Rack 1: 仅 LD 29 (1个LD), Cluster 5 不存在
// Rack 2: LD 36-53 完整
// 13 LD 从 LD 6 会跨越到 Rack 1 → 应被拒绝 → 应从 Rack 2 开始
func TestAllocate_13LDs_NoCrossRack(t *testing.T) {
	var avail []string

	// Rack 1: LD 29
	for d := 0; d < domainsPerLD; d++ {
		avail = append(avail, fmt.Sprintf("%d.%d", 29, d))
	}

	// Rack 0: LD 0 (D1-D7), LD 1 (D4-D7), LD 2-17 (完整)
	for d := 1; d < domainsPerLD; d++ {
		avail = append(avail, fmt.Sprintf("%d.%d", 0, d))
	}
	for d := 4; d < domainsPerLD; d++ {
		avail = append(avail, fmt.Sprintf("%d.%d", 1, d))
	}
	for ld := 2; ld <= 17; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}

	// Rack 2: LD 36-53 完整
	for ld := 36; ld <= 53; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}

	// 申请 104 Domain = 13 LD (>6, <=18)
	result, racks, err := Allocate(104, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("result[0]=%s, racks=%v", result[0], racks)

	if result[0] != "36.0" {
		t.Errorf("expected start from LD 36 (Rack 2), got %s", result[0])
	}
	if racks[0] != 2 {
		t.Errorf("expected Rack 2, got %v", racks)
	}
}

// 验证 <=18 LD 不能跨Rack 是硬约束
func TestAllocate_NoCrossRackWhenWithinCapacity(t *testing.T) {
	// Rack 0: LD 12-17 (Cluster 2 only)
	// Rack 1: LD 18-23 (Cluster 3 only)
	// 10 LD (>6, <=18): LD 12起始跨Rack → 拒绝
	// LD 18起始需要 LD 18-27 → LD 24-27不存在 → 失败
	var avail []string
	for ld := 12; ld < 18; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	for ld := 18; ld < 24; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}

	_, _, err := Allocate(80, avail, PZ1)
	if err == nil {
		t.Fatal("expected error: cannot fit 10 LDs in a single rack")
	}
	t.Logf("expected error: %v", err)
}
