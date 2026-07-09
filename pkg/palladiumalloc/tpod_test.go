package palladiumalloc

import (
	"reflect"
	"testing"
)

func TestAllocateTPods_SingleRack(t *testing.T) {
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

	result, err := AllocateTPods([]int{0}, tpodReqs, availableTPods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []AllocatedTPod{
		{RackId: 0, TPodId: 0, ExtType: "USB-HDSB"},
		{RackId: 0, TPodId: 1, ExtType: "USB-HDSB"},
		{RackId: 0, TPodId: 3, ExtType: "PCI"},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocateTPods_SecondRackUsed(t *testing.T) {
	tpodReqs := []TPodRequirement{
		{ExtType: "USB-HDSB", Number: 2},
	}
	availableTPods := map[int]map[string][]int{
		0: {"PCI": {0, 1}},                     // 没有USB-HDSB
		1: {"USB-HDSB": {0, 1, 2}},             // 有USB-HDSB
	}

	result, err := AllocateTPods([]int{0, 1}, tpodReqs, availableTPods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 应该从 Rack 1 分配
	if result[0].RackId != 1 {
		t.Errorf("expected Rack 1, got Rack %d", result[0].RackId)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 TPods, got %d", len(result))
	}
}

func TestAllocateTPods_NoRackSatisfies(t *testing.T) {
	tpodReqs := []TPodRequirement{
		{ExtType: "USB-HDSB", Number: 5}, // 需要5个
	}
	availableTPods := map[int]map[string][]int{
		0: {"USB-HDSB": {0, 1, 2}}, // 只有3个
	}

	_, err := AllocateTPods([]int{0}, tpodReqs, availableTPods)
	if err == nil {
		t.Fatal("expected error: insufficient TPod quantity")
	}
}

func TestAllocateTPods_EmptyReqs(t *testing.T) {
	_, err := AllocateTPods([]int{0}, nil, map[int]map[string][]int{0: {}})
	if err == nil {
		t.Fatal("expected error: empty requirements")
	}
}

func TestAllocateTPods_EmptyRackIds(t *testing.T) {
	tpodReqs := []TPodRequirement{{ExtType: "USB", Number: 1}}
	_, err := AllocateTPods([]int{}, tpodReqs, map[int]map[string][]int{0: {"USB": {0}}})
	if err == nil {
		t.Fatal("expected error: empty rack IDs")
	}
}

func TestAllocateTPods_NoAvailableTPods(t *testing.T) {
	tpodReqs := []TPodRequirement{{ExtType: "USB", Number: 1}}
	_, err := AllocateTPods([]int{0}, tpodReqs, nil)
	if err == nil {
		t.Fatal("expected error: no available TPods")
	}
}
