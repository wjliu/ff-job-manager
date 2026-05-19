package zebualloc

import (
	"fmt"
	"reflect"
	"testing"
)

// 生成zs3/zs4的完整Unit可用列表
func makeHalfModuleAvailable(unitCount int) []string {
	var avail []string
	for u := 0; u < unitCount; u++ {
		for hm := 0; hm < 8; hm++ {
			avail = append(avail, fmt.Sprintf("U%d.HM%d", u, hm))
		}
	}
	return avail
}

// 生成zs5的完整Unit可用列表
func makeSubModuleAvailable(unitCount int) []string {
	var avail []string
	for u := 0; u < unitCount; u++ {
		for m := 0; m < 4; m++ {
			for s := 0; s < 4; s++ {
				avail = append(avail, fmt.Sprintf("U%d.M%d.S%d", u, m, s))
			}
		}
	}
	return avail
}

func TestAllocate_HalfModule_LessThan4Modules(t *testing.T) {
	// 申请2个HalfModule = 1个Module, <4规则: 必须从M0开始
	avail := makeHalfModuleAvailable(2) // U0, U1
	result, err := Allocate(2, avail, ZS3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"U0.HM0", "U0.HM1"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_HalfModule_Exactly4Modules(t *testing.T) {
	// 申请8个HalfModule = 4个Module = 1个完整Unit
	avail := makeHalfModuleAvailable(2)
	result, err := Allocate(8, avail, ZS4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{
		"U0.HM0", "U0.HM1", "U0.HM2", "U0.HM3",
		"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_HalfModule_5Modules(t *testing.T) {
	// 申请10个HalfModule = 5个Module = 1个完整Unit + 1个Module
	avail := makeHalfModuleAvailable(2)
	result, err := Allocate(10, avail, ZS3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1个完整Unit(U0) + 1个Module(U1.M0)
	expected := []string{
		"U0.HM0", "U0.HM1", "U0.HM2", "U0.HM3",
		"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7",
		"U1.HM0", "U1.HM1",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_SubModule_LessThan4Modules(t *testing.T) {
	// 申请4个SubModule = 1个Module, <4规则: 必须从M0开始
	avail := makeSubModuleAvailable(2)
	result, err := Allocate(4, avail, ZS5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"U0.M0.S0", "U0.M0.S1", "U0.M0.S2", "U0.M0.S3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_SubModule_Exactly4Modules(t *testing.T) {
	// 申请16个SubModule = 4个Module = 1个完整Unit
	avail := makeSubModuleAvailable(2)
	result, err := Allocate(16, avail, ZS5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{
		"U0.M0.S0", "U0.M0.S1", "U0.M0.S2", "U0.M0.S3",
		"U0.M1.S0", "U0.M1.S1", "U0.M1.S2", "U0.M1.S3",
		"U0.M2.S0", "U0.M2.S1", "U0.M2.S2", "U0.M2.S3",
		"U0.M3.S0", "U0.M3.S1", "U0.M3.S2", "U0.M3.S3",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_SubModule_5Modules(t *testing.T) {
	// 申请20个SubModule = 5个Module = 1个完整Unit + 1个Module
	avail := makeSubModuleAvailable(2)
	result, err := Allocate(20, avail, ZS5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1个完整Unit(U0) + 1个Module(U1.M0)
	expected := []string{
		"U0.M0.S0", "U0.M0.S1", "U0.M0.S2", "U0.M0.S3",
		"U0.M1.S0", "U0.M1.S1", "U0.M1.S2", "U0.M1.S3",
		"U0.M2.S0", "U0.M2.S1", "U0.M2.S2", "U0.M2.S3",
		"U0.M3.S0", "U0.M3.S1", "U0.M3.S2", "U0.M3.S3",
		"U1.M0.S0", "U1.M0.S1", "U1.M0.S2", "U1.M0.S3",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_UnitSkip(t *testing.T) {
	// U0不可用(不提供), 只有U1和U3可用 — Unit间可跳选
	avail := []string{
		// U1完整
		"U1.HM0", "U1.HM1", "U1.HM2", "U1.HM3",
		"U1.HM4", "U1.HM5", "U1.HM6", "U1.HM7",
		// U3完整
		"U3.HM0", "U3.HM1", "U3.HM2", "U3.HM3",
		"U3.HM4", "U3.HM5", "U3.HM6", "U3.HM7",
	}
	result, err := Allocate(8, avail, ZS3) // 4 Modules = 1 complete Unit
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{
		"U1.HM0", "U1.HM1", "U1.HM2", "U1.HM3",
		"U1.HM4", "U1.HM5", "U1.HM6", "U1.HM7",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_InsufficientModules(t *testing.T) {
	// 只有1个完整Unit，但申请2个完整Unit
	avail := makeHalfModuleAvailable(1)
	_, err := Allocate(16, avail, ZS3) // 8 Modules = 2 Units
	if err == nil {
		t.Fatal("expected error for insufficient modules")
	}
}

func TestAllocate_ZeroCount(t *testing.T) {
	avail := makeHalfModuleAvailable(1)
	_, err := Allocate(0, avail, ZS3)
	if err == nil {
		t.Fatal("expected error for zero count")
	}
}

func TestAllocate_EmptyAvailable(t *testing.T) {
	_, err := Allocate(2, []string{}, ZS3)
	if err == nil {
		t.Fatal("expected error for empty available list")
	}
}

func TestAllocate_InvalidHalfModuleName(t *testing.T) {
	avail := []string{"U0.HM0", "INVALID"}
	_, err := Allocate(2, avail, ZS3)
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
}

func TestAllocate_InvalidSubModuleName(t *testing.T) {
	avail := []string{"U0.M0.S0", "BAD_NAME"}
	_, err := Allocate(4, avail, ZS5)
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
}

func TestAllocate_IncompleteModuleNotUsed(t *testing.T) {
	// U0.M0不完整(只有HM0), U0.M1完整, U1.M0完整
	avail := []string{
		"U0.HM0",          // U0.M0不完整, 缺少HM1
		"U0.HM2", "U0.HM3", // U0.M1完整
		"U1.HM0", "U1.HM1", // U1.M0完整
	}
	// 申请2个HalfModule = 1个Module, <4规则: 不要求从M0开始
	// U0.M0不完整不可用, U0.M1完整可用（<4不要求M0起），优先选U0.M1
	result, err := Allocate(2, avail, ZS3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"U0.HM2", "U0.HM3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_RemainingFromNonM0(t *testing.T) {
	// 5个Module: 1个完整Unit + 1个Module
	// 剩余1个Module可以在任意Unit内从任意位置开始（>=4规则下的剩余部分）
	avail := []string{
		// U0完整
		"U0.HM0", "U0.HM1", "U0.HM2", "U0.HM3",
		"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7",
		// U1只有M1可用
		"U1.HM2", "U1.HM3",
	}
	result, err := Allocate(10, avail, ZS3) // 5 Modules
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// U0完整 + U1.M1
	expected := []string{
		"U0.HM0", "U0.HM1", "U0.HM2", "U0.HM3",
		"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7",
		"U1.HM2", "U1.HM3",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_LessThan4NotFromM0(t *testing.T) {
	// <4规则: 不要求从M0开始，只要Unit内连续即可
	// U0只有M1可用(M0不可用)
	avail := []string{
		"U0.HM2", "U0.HM3", // U0.M1
	}
	result, err := Allocate(2, avail, ZS3) // 1 Module, <4规则
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 应该选U0.M1（不从M0开始也可以）
	expected := []string{"U0.HM2", "U0.HM3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestParseSystemType(t *testing.T) {
	tests := []struct {
		input string
		want  SystemType
	}{
		{"zs3", ZS3},
		{"ZS3", ZS3},
		{"zs4", ZS4},
		{"ZS4", ZS4},
		{"zs5", ZS5},
		{"ZS5", ZS5},
	}
	for _, tt := range tests {
		got, err := ParseSystemType(tt.input)
		if err != nil {
			t.Errorf("ParseSystemType(%q): unexpected error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("ParseSystemType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}

	_, err := ParseSystemType("unknown")
	if err == nil {
		t.Error("expected error for unknown system type")
	}
}

func TestSystemTypeString(t *testing.T) {
	if ZS3.String() != "zs3" {
		t.Errorf("ZS3.String() = %q, want %q", ZS3.String(), "zs3")
	}
	if ZS4.String() != "zs4" {
		t.Errorf("ZS4.String() = %q, want %q", ZS4.String(), "zs4")
	}
	if ZS5.String() != "zs5" {
		t.Errorf("ZS5.String() = %q, want %q", ZS5.String(), "zs5")
	}
}

func TestAllocate_WrapAround_M3ToM0(t *testing.T) {
	// 回转规则: M3和M0视为连续，M3→M0回转分配
	// U0只有M3和M0可用（M1,M2不可用）
	avail := []string{
		"U0.HM6", "U0.HM7", // U0.M3
		"U0.HM0", "U0.HM1", // U0.M0
	}
	// 申请2个Module，<4规则，M3→M0回转连续
	result, err := Allocate(4, avail, ZS3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"U0.HM6", "U0.HM7", "U0.HM0", "U0.HM1"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_WrapAround_ThreeModulesM2M3M0(t *testing.T) {
	// 回转: M2→M3→M0 三连
	avail := []string{
		"U0.HM4", "U0.HM5", // U0.M2
		"U0.HM6", "U0.HM7", // U0.M3
		"U0.HM0", "U0.HM1", // U0.M0
	}
	// 申请3个Module (6 HalfModule)
	result, err := Allocate(6, avail, ZS3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7", "U0.HM0", "U0.HM1"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_WrapAround_ZS5(t *testing.T) {
	// zs5回转: M3→M0
	avail := []string{
		"U0.M3.S0", "U0.M3.S1", "U0.M3.S2", "U0.M3.S3", // U0.M3
		"U0.M0.S0", "U0.M0.S1", "U0.M0.S2", "U0.M0.S3", // U0.M0
	}
	// 申请2个Module (8 SubModule)
	result, err := Allocate(8, avail, ZS5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{
		"U0.M3.S0", "U0.M3.S1", "U0.M3.S2", "U0.M3.S3",
		"U0.M0.S0", "U0.M0.S1", "U0.M0.S2", "U0.M0.S3",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_WrapAround_RemainingAfterCompleteUnit(t *testing.T) {
	// >=4规则: 1个完整Unit + 剩余1个Module回转分配
	// U1只有M3和M0可用，剩余1个Module可从M3或M0开始
	avail := []string{
		// U0完整
		"U0.HM0", "U0.HM1", "U0.HM2", "U0.HM3",
		"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7",
		// U1只有M3和M0可用
		"U1.HM6", "U1.HM7", // U1.M3
		"U1.HM0", "U1.HM1", // U1.M0
	}
	// 申请5 Module (10 HalfModule): 1完整Unit + 1剩余
	result, err := Allocate(10, avail, ZS3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 完整Unit U0 + U1.M0 (M0在排序中优先于M3作为起点)
	expected := []string{
		"U0.HM0", "U0.HM1", "U0.HM2", "U0.HM3",
		"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7",
		"U1.HM0", "U1.HM1",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAllocate_WrapAround_NotEnoughConsecutive(t *testing.T) {
	// 回转也无法满足: U0只有M1和M3可用，M1和M3不连续（M1→M2→M3缺M2，M3→M0→M1缺M0）
	avail := []string{
		"U0.HM2", "U0.HM3", // U0.M1
		"U0.HM6", "U0.HM7", // U0.M3
	}
	// 申请2个Module (4 HalfModule)
	_, err := Allocate(4, avail, ZS3)
	if err == nil {
		t.Fatal("expected error for non-consecutive modules")
	}
}

func TestAllocate_WrapAround_M3OnlyStart(t *testing.T) {
	// M3是唯一能开始连续分配2个Module的起点: M3→M0回转
	// M0本身不可用，只有M3和M0可用
	avail := []string{
		"U0.HM6", "U0.HM7", // U0.M3
		"U0.M0.S0", // 不合法格式，应该被忽略 — 但这里测试U0只有M3
	}
	// U0只有M3可用, 无法分配2个连续Module
	_, err := Allocate(4, avail, ZS3)
	if err == nil {
		t.Fatal("expected error: only M3 available, need 2 consecutive modules")
	}
}

func TestAllocate_WrapAround_M3AndM0Only(t *testing.T) {
	// U0只有M3和M0可用，M3→M0回转，申请2个Module
	avail := []string{
		"U0.HM6", "U0.HM7", // U0.M3
		"U0.HM0", "U0.HM1", // U0.M0
	}
	result, err := Allocate(4, avail, ZS3) // 2 Modules
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 从M0开始分配: M0, M1(不可用) — 不行
	// 从M3开始回转: M3, M0 — 连续
	expected := []string{"U0.HM6", "U0.HM7", "U0.HM0", "U0.HM1"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
