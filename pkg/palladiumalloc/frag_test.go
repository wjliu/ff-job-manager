package palladiumalloc

import (
	"fmt"
	"reflect"
	"testing"
)

// 规则3碎片测试: 已有部分使用的Rack优先于完全空闲的Rack
func TestAllocate_FragmentationPrefersPartiallyUsedRack(t *testing.T) {
	// Rack 0: Cluster 0 (LD 0-5) 已被占用，Cluster 1(LD 6-11) 和 Cluster 2(LD 12-17) 空闲
	// Rack 1: LD 18-35 全部空闲
	// 期望: 优先分配到 Rack 0（已用LD多，碎片多，应优先填满）
	var avail []string
	for ld := 6; ld < 18; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	for ld := 18; ld < 36; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}

	// 申请48 Domain = 6 LD (neededLDs=6, <=6)
	// Rack 0: 12个可用LD 优先于 Rack 1: 18个可用LD
	result, racks, err := Allocate(48, avail, PZ1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 48 {
		t.Errorf("expected 48 domains, got %d", len(result))
	}
	if result[0] != "6.0" {
		t.Errorf("expected start from LD 6 (partially used Rack 0), got %s", result[0])
	}
	if !reflect.DeepEqual(racks, []int{0}) {
		t.Errorf("expected racks [0], got %v", racks)
	}
}
