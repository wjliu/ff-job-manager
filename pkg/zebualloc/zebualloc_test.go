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
		"U0.HM0",           // U0.M0不完整, 缺少HM1
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

func TestAllocate_ZS3_NonMultipleCount(t *testing.T) {
	avail := makeHalfModuleAvailable(2)

	// 申请1个HalfModule
	result, err := Allocate(1, avail, ZS3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 HalfModule, got %d: %v", len(result), result)
	}
	if result[0] != "U0.HM0" {
		t.Errorf("expected U0.HM0, got %s", result[0])
	}

	// 申请3个HalfModule
	result, err = Allocate(3, avail, ZS3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 HalfModules, got %d: %v", len(result), result)
	}
	expected3 := []string{"U0.HM0", "U0.HM1", "U0.HM2"}
	if !reflect.DeepEqual(result, expected3) {
		t.Errorf("expected %v, got %v", expected3, result)
	}

	// 申请5个HalfModule
	result, err = Allocate(5, avail, ZS3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("expected 5 HalfModules, got %d: %v", len(result), result)
	}
}

func TestAllocate_ZS5_NonMultipleCount(t *testing.T) {
	avail := makeSubModuleAvailable(2)

	// 申请1个SubModule
	result, err := Allocate(1, avail, ZS5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 SubModule, got %d: %v", len(result), result)
	}
	if result[0] != "U0.M0.S0" {
		t.Errorf("expected U0.M0.S0, got %s", result[0])
	}

	// 申请3个SubModule
	result, err = Allocate(3, avail, ZS5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 SubModules, got %d: %v", len(result), result)
	}
	expected3 := []string{"U0.M0.S0", "U0.M0.S1", "U0.M0.S2"}
	if !reflect.DeepEqual(result, expected3) {
		t.Errorf("expected %v, got %v", expected3, result)
	}

	// 申请6个SubModule
	result, err = Allocate(6, avail, ZS5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 6 {
		t.Errorf("expected 6 SubModules, got %d: %v", len(result), result)
	}
}

// ========== FormatEnvVar 测试 ==========

func TestFormatEnvVar_Empty(t *testing.T) {
	result := FormatEnvVar([]string{}, ZS3)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestFormatEnvVar_SingleModule_ZS3(t *testing.T) {
	// 分配2个HalfModule = 1个完整Module
	allocated := []string{"U0.HM0", "U0.HM1"}
	result := FormatEnvVar(allocated, ZS3)
	expected := "U0.M0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_CompleteUnit_ZS3(t *testing.T) {
	// 分配8个HalfModule = 4个完整Module = 1个完整Unit
	allocated := []string{
		"U0.HM0", "U0.HM1", "U0.HM2", "U0.HM3",
		"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7",
	}
	result := FormatEnvVar(allocated, ZS3)
	expected := "U0.M0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_CompleteUnit_ZS5(t *testing.T) {
	// 分配16个SubModule = 4个完整Module = 1个完整Unit
	allocated := []string{
		"U0.M0.S0", "U0.M0.S1", "U0.M0.S2", "U0.M0.S3",
		"U0.M1.S0", "U0.M1.S1", "U0.M1.S2", "U0.M1.S3",
		"U0.M2.S0", "U0.M2.S1", "U0.M2.S2", "U0.M2.S3",
		"U0.M3.S0", "U0.M3.S1", "U0.M3.S2", "U0.M3.S3",
	}
	result := FormatEnvVar(allocated, ZS5)
	expected := "U0.M0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_CompleteUnitPlusPartial(t *testing.T) {
	// 分配: U1完整Unit, U0部分Module → spec rule 4 example
	allocated := []string{
		// U1: 完整 Unit (M0-M3)
		"U1.HM0", "U1.HM1", "U1.HM2", "U1.HM3",
		"U1.HM4", "U1.HM5", "U1.HM6", "U1.HM7",
		// U0: 部分 Unit (M2, M3)
		"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7",
	}
	result := FormatEnvVar(allocated, ZS3)
	// 完整Unit U1.M0 在前, 部分Unit U0.M2 在后
	expected := "U1.M0,U0.M2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_MultipleCompleteUnits(t *testing.T) {
	// 分配2个完整Unit, U1在前(先出现), U0在后
	allocated := []string{
		"U1.HM0", "U1.HM1", "U1.HM2", "U1.HM3",
		"U1.HM4", "U1.HM5", "U1.HM6", "U1.HM7",
		"U0.HM0", "U0.HM1", "U0.HM2", "U0.HM3",
		"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7",
	}
	result := FormatEnvVar(allocated, ZS3)
	expected := "U1.M0,U0.M0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_PartialUnit_NonZeroStart(t *testing.T) {
	// 部分Unit从M2开始分配
	allocated := []string{
		"U0.HM4", "U0.HM5", // U0.M2
		"U0.HM6", "U0.HM7", // U0.M3
	}
	result := FormatEnvVar(allocated, ZS3)
	expected := "U0.M2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_WrapAround_PreserveFirstModule(t *testing.T) {
	// 回转分配: U0.M3 → U0.M0, M3是第一个分配的Module
	allocated := []string{
		"U0.HM6", "U0.HM7", // U0.M3 (first)
		"U0.HM0", "U0.HM1", // U0.M0
	}
	result := FormatEnvVar(allocated, ZS3)
	// M3是第一个分配的Module, 注入U0.M3而不是U0.M0 (spec rule 3)
	expected := "U0.M3"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_ZS5_PartialUnit(t *testing.T) {
	// zs5 部分Unit: M1, M2
	allocated := []string{
		"U0.M1.S0", "U0.M1.S1", "U0.M1.S2", "U0.M1.S3",
		"U0.M2.S0", "U0.M2.S1", "U0.M2.S2", "U0.M2.S3",
	}
	result := FormatEnvVar(allocated, ZS5)
	expected := "U0.M1"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_IncompleteModule_ZS3(t *testing.T) {
	// 分配3个HalfModule: M0完整, M1不完整(只有HM2)
	allocated := []string{
		"U0.HM0", "U0.HM1", // U0.M0 完整
		"U0.HM2", // U0.M1 不完整(缺少HM3)
	}
	result := FormatEnvVar(allocated, ZS3)
	// M0合并为U0.M0, 不完整的HM2单独列出
	expected := "U0.M0,U0.HM2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_IncompleteModule_ZS5(t *testing.T) {
	// 分配6个SubModule: M0完整, M1不完整(只有S0,S1)
	allocated := []string{
		"U0.M0.S0", "U0.M0.S1", "U0.M0.S2", "U0.M0.S3", // U0.M0 完整
		"U0.M1.S0", "U0.M1.S1", // U0.M1 不完整(缺少S2,S3)
	}
	result := FormatEnvVar(allocated, ZS5)
	expected := "U0.M0,U0.M1.S0,U0.M1.S1"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_OnlyIncompleteModule_ZS3(t *testing.T) {
	// 只分配1个HalfModule: 不完整Module
	allocated := []string{"U0.HM0"}
	result := FormatEnvVar(allocated, ZS3)
	expected := "U0.HM0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_OnlyIncompleteModule_ZS5(t *testing.T) {
	// 只分配1个SubModule: 不完整Module
	allocated := []string{"U0.M0.S0"}
	result := FormatEnvVar(allocated, ZS5)
	expected := "U0.M0.S0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_AllIncompleteModules_ZS5(t *testing.T) {
	// 分配3个SubModule: M0不完整(只有S0,S1,S2)
	allocated := []string{
		"U0.M0.S0", "U0.M0.S1", "U0.M0.S2",
	}
	result := FormatEnvVar(allocated, ZS5)
	expected := "U0.M0.S0,U0.M0.S1,U0.M0.S2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_MultipleUnits_MixedPartial(t *testing.T) {
	// 两个部分Unit: U0.M2, U0.M3 和 U2.M0
	allocated := []string{
		"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7", // U0.M2, U0.M3
		"U2.HM0", "U2.HM1", // U2.M0
	}
	result := FormatEnvVar(allocated, ZS3)
	// 按首次出现顺序: U0先, U2后
	expected := "U0.M2,U2.M0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_EndToEnd_ZS3_10HalfModules(t *testing.T) {
	// 端到端测试: Allocate → FormatEnvVar
	avail := makeHalfModuleAvailable(2)
	allocated, err := Allocate(10, avail, ZS3) // 5 Modules = 1完整Unit + 1部分Module
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := FormatEnvVar(allocated, ZS3)
	// U0完整 → U0.M0, U1.M0部分 → U1.M0
	expected := "U0.M0,U1.M0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_EndToEnd_ZS5_20SubModules(t *testing.T) {
	// 端到端测试: Allocate → FormatEnvVar
	avail := makeSubModuleAvailable(2)
	allocated, err := Allocate(20, avail, ZS5) // 5 Modules = 1完整Unit + 1部分Module
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := FormatEnvVar(allocated, ZS5)
	expected := "U0.M0,U1.M0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_EndToEnd_ZS3_NonMultiple(t *testing.T) {
	// 端到端: 3个HalfModule (非倍数)
	avail := makeHalfModuleAvailable(2)
	allocated, err := Allocate(3, avail, ZS3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// allocated = [U0.HM0, U0.HM1, U0.HM2]
	result := FormatEnvVar(allocated, ZS3)
	// U0.M0完整, U0.M1不完整(只有HM2)
	expected := "U0.M0,U0.HM2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_EndToEnd_ZS5_NonMultiple(t *testing.T) {
	// 端到端: 6个SubModule (非倍数)
	avail := makeSubModuleAvailable(2)
	allocated, err := Allocate(6, avail, ZS5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// allocated = [U0.M0.S0-S3, U0.M1.S0-S1]
	result := FormatEnvVar(allocated, ZS5)
	expected := "U0.M0,U0.M1.S0,U0.M1.S1"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_EndToEnd_WrapAround(t *testing.T) {
	// 端到端: 回转分配 M3→M0
	avail := []string{
		"U0.HM6", "U0.HM7", // U0.M3
		"U0.HM0", "U0.HM1", // U0.M0
	}
	allocated, err := Allocate(4, avail, ZS3) // 2 Modules
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := FormatEnvVar(allocated, ZS3)
	// 分配从M3开始回转, M3在分配顺序中先出现, 所以first是M3 (spec rule 3)
	expected := "U0.M3"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_ZS4(t *testing.T) {
	// zs4与zs3使用相同的HalfModule格式
	allocated := []string{
		"U0.HM0", "U0.HM1", "U0.HM2", "U0.HM3",
		"U0.HM4", "U0.HM5", "U0.HM6", "U0.HM7",
	}
	result := FormatEnvVar(allocated, ZS4)
	expected := "U0.M0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_UserScenario_CompleteUnitFirst(t *testing.T) {
	// 复现用户场景: zs4机型, U7.M2, U7.M3, U9.M0, U9.M1, U9.M2, U9.M3
	// 期望: 完整Unit U9.M0 在前, 部分Unit U7.M2 在后
	avail := []string{
		// U7: 只有 M2, M3
		"U7.HM4", "U7.HM5", // U7.M2
		"U7.HM6", "U7.HM7", // U7.M3
		// U9: 完整 Unit (M0-M3)
		"U9.HM0", "U9.HM1", // U9.M0
		"U9.HM2", "U9.HM3", // U9.M1
		"U9.HM4", "U9.HM5", // U9.M2
		"U9.HM6", "U9.HM7", // U9.M3
	}

	// 申请 12 个 HalfModule = 6 个 Module (>=4 规则: 1 完整 Unit + 2 剩余)
	allocated, err := Allocate(12, avail, ZS4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("allocated devices: %v", allocated)

	result := FormatEnvVar(allocated, ZS4)
	t.Logf("FormatEnvVar result: %s", result)

	// 完整 Unit U9.M0 应在前, 部分 Unit U7.M2 应在后
	expected := "U9.M0,U7.M2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_UserScenario_U7FirstInInput(t *testing.T) {
	// 即使 U7 在输入中先出现, 完整 Unit 也应在前面
	allocated := []string{
		"U7.HM4", "U7.HM5", "U7.HM6", "U7.HM7",
		"U9.HM0", "U9.HM1", "U9.HM2", "U9.HM3",
		"U9.HM4", "U9.HM5", "U9.HM6", "U9.HM7",
	}

	result := FormatEnvVar(allocated, ZS4)
	t.Logf("FormatEnvVar result: %s", result)

	expected := "U9.M0,U7.M2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatEnvVar_WrapAround_PartialUnit_M3AndM0(t *testing.T) {
	// 复现: zs4, U7.M0 + U7.M3 (回转), U9.M0-M3 (完整)
	// 预期: U9.M0,U7.M3 — M3是回转的首个Module (spec rule 3)
	avail := []string{
		// U7: M0 和 M3 (回转场景)
		"U7.HM0", "U7.HM1", // U7.M0
		"U7.HM6", "U7.HM7", // U7.M3
		// U9: 完整 Unit
		"U9.HM0", "U9.HM1", // U9.M0
		"U9.HM2", "U9.HM3", // U9.M1
		"U9.HM4", "U9.HM5", // U9.M2
		"U9.HM6", "U9.HM7", // U9.M3
	}

	// 申请 12 个 HalfModule = 6 个 Module: 1 完整 Unit + 2 剩余 (M3→M0 回转)
	allocated, err := Allocate(12, avail, ZS4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("allocated devices: %v", allocated)

	result := FormatEnvVar(allocated, ZS4)
	t.Logf("FormatEnvVar result: %s", result)

	// 完整 Unit U9.M0 在前, 部分 Unit U7.M3 (回转首个Module) 在后
	expected := "U9.M0,U7.M3"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
