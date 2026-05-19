package zebualloc

import (
	"fmt"
	"testing"
)

// generateHalfModuleList 生成指定Unit数量的HalfModule可用列表
func generateHalfModuleList(unitCount int) []string {
	avail := make([]string, 0, unitCount*8)
	for u := 0; u < unitCount; u++ {
		for hm := 0; hm < 8; hm++ {
			avail = append(avail, fmt.Sprintf("U%d.HM%d", u, hm))
		}
	}
	return avail
}

// generateSubModuleList 生成指定Unit数量的SubModule可用列表
func generateSubModuleList(unitCount int) []string {
	avail := make([]string, 0, unitCount*16)
	for u := 0; u < unitCount; u++ {
		for m := 0; m < 4; m++ {
			for s := 0; s < 4; s++ {
				avail = append(avail, fmt.Sprintf("U%d.M%d.S%d", u, m, s))
			}
		}
	}
	return avail
}

// Benchmark 分配场景: ZS3 HalfModule 小规模请求 (< 4 Modules)
func BenchmarkAllocate_ZS3_SmallRequest(b *testing.B) {
	avail := generateHalfModuleList(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(2, avail, ZS3)
	}
}

// Benchmark 分配场景: ZS3 HalfModule 恰好1个完整Unit
func BenchmarkAllocate_ZS3_OneCompleteUnit(b *testing.B) {
	avail := generateHalfModuleList(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(8, avail, ZS3)
	}
}

// Benchmark 分配场景: ZS3 HalfModule 1个完整Unit + 剩余
func BenchmarkAllocate_ZS3_CompleteUnitPlusRemaining(b *testing.B) {
	avail := generateHalfModuleList(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(10, avail, ZS3)
	}
}

// Benchmark 分配场景: ZS3 HalfModule 多个完整Unit
func BenchmarkAllocate_ZS3_MultipleCompleteUnits(b *testing.B) {
	avail := generateHalfModuleList(16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(32, avail, ZS3)
	}
}

// Benchmark 分配场景: ZS4 HalfModule 小规模请求
func BenchmarkAllocate_ZS4_SmallRequest(b *testing.B) {
	avail := generateHalfModuleList(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(2, avail, ZS4)
	}
}

// Benchmark 分配场景: ZS4 HalfModule 恰好1个完整Unit
func BenchmarkAllocate_ZS4_OneCompleteUnit(b *testing.B) {
	avail := generateHalfModuleList(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(8, avail, ZS4)
	}
}

// Benchmark 分配场景: ZS5 SubModule 小规模请求 (< 4 Modules)
func BenchmarkAllocate_ZS5_SmallRequest(b *testing.B) {
	avail := generateSubModuleList(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(4, avail, ZS5)
	}
}

// Benchmark 分配场景: ZS5 SubModule 恰好1个完整Unit
func BenchmarkAllocate_ZS5_OneCompleteUnit(b *testing.B) {
	avail := generateSubModuleList(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(16, avail, ZS5)
	}
}

// Benchmark 分配场景: ZS5 SubModule 1个完整Unit + 剩余
func BenchmarkAllocate_ZS5_CompleteUnitPlusRemaining(b *testing.B) {
	avail := generateSubModuleList(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(20, avail, ZS5)
	}
}

// Benchmark 分配场景: ZS5 SubModule 多个完整Unit
func BenchmarkAllocate_ZS5_MultipleCompleteUnits(b *testing.B) {
	avail := generateSubModuleList(16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(64, avail, ZS5)
	}
}

// Benchmark: 不同规模的可用设备列表对分配性能的影响
func BenchmarkAllocate_ZS3_Scale_4Units(b *testing.B)   { benchZS3Scale(b, 4) }
func BenchmarkAllocate_ZS3_Scale_16Units(b *testing.B)  { benchZS3Scale(b, 16) }
func BenchmarkAllocate_ZS3_Scale_64Units(b *testing.B)  { benchZS3Scale(b, 64) }
func BenchmarkAllocate_ZS3_Scale_256Units(b *testing.B) { benchZS3Scale(b, 256) }

func benchZS3Scale(b *testing.B, unitCount int) {
	avail := generateHalfModuleList(unitCount)
	// 申请1个Module (2 HalfModule), 测试buildUnitMap在不同规模下的开销
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(2, avail, ZS3)
	}
}

func BenchmarkAllocate_ZS5_Scale_4Units(b *testing.B)   { benchZS5Scale(b, 4) }
func BenchmarkAllocate_ZS5_Scale_16Units(b *testing.B)  { benchZS5Scale(b, 16) }
func BenchmarkAllocate_ZS5_Scale_64Units(b *testing.B)  { benchZS5Scale(b, 64) }
func BenchmarkAllocate_ZS5_Scale_256Units(b *testing.B) { benchZS5Scale(b, 256) }

func benchZS5Scale(b *testing.B, unitCount int) {
	avail := generateSubModuleList(unitCount)
	// 申请1个Module (4 SubModule), 测试buildUnitMap在不同规模下的开销
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(4, avail, ZS5)
	}
}

// Benchmark: 大规模分配请求
func BenchmarkAllocate_ZS3_LargeRequest(b *testing.B) {
	avail := generateHalfModuleList(64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(256, avail, ZS3) // 128 Modules = 32 完整Unit
	}
}

func BenchmarkAllocate_ZS5_LargeRequest(b *testing.B) {
	avail := generateSubModuleList(64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(512, avail, ZS5) // 128 Modules = 32 完整Unit
	}
}

// Benchmark: parseHalfModule 解析性能
func BenchmarkParseHalfModule(b *testing.B) {
	names := []string{"U0.HM0", "U3.HM5", "U15.HM7", "U99.HM1"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, n := range names {
			parseHalfModule(n)
		}
	}
}

// Benchmark: parseSubModule 解析性能
func BenchmarkParseSubModule(b *testing.B) {
	names := []string{"U0.M0.S0", "U3.M2.S1", "U15.M3.S3", "U99.M1.S0"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, n := range names {
			parseSubModule(n)
		}
	}
}

// Benchmark: buildUnitMap 构建性能 (ZS3)
func BenchmarkBuildUnitMap_ZS3(b *testing.B) {
	avail := generateHalfModuleList(16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildUnitMap(avail, ZS3)
	}
}

// Benchmark: buildUnitMap 构建性能 (ZS5)
func BenchmarkBuildUnitMap_ZS5(b *testing.B) {
	avail := generateSubModuleList(16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildUnitMap(avail, ZS5)
	}
}

// Benchmark: modulesToNames 转换性能
func BenchmarkModulesToNames_ZS3(b *testing.B) {
	mods := make([]moduleUnit, 16) // 16 Modules = 4 完整Unit
	for i := range mods {
		mods[i] = moduleUnit{unitIndex: i / 4, moduleIndex: i % 4}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		modulesToNames(mods, ZS3)
	}
}

func BenchmarkModulesToNames_ZS5(b *testing.B) {
	mods := make([]moduleUnit, 16)
	for i := range mods {
		mods[i] = moduleUnit{unitIndex: i / 4, moduleIndex: i % 4}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		modulesToNames(mods, ZS5)
	}
}

// Benchmark: 稀疏可用设备 (部分Unit可用) 场景
func BenchmarkAllocate_ZS3_SparseAvailable(b *testing.B) {
	// 模拟只有偶数Unit可用的场景
	var avail []string
	for u := 0; u < 64; u += 2 {
		for hm := 0; hm < 8; hm++ {
			avail = append(avail, fmt.Sprintf("U%d.HM%d", u, hm))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(8, avail, ZS3)
	}
}

// Benchmark: 不完整Module (需过滤) 场景
func BenchmarkAllocate_ZS5_IncompleteModules(b *testing.B) {
	// 每个Unit部分Module不完整，测试过滤逻辑的开销
	var avail []string
	for u := 0; u < 32; u++ {
		for m := 0; m < 4; m++ {
			// 奇数Module缺少一个SubModule，不完整
			if m%2 == 1 {
				for s := 0; s < 3; s++ {
					avail = append(avail, fmt.Sprintf("U%d.M%d.S%d", u, m, s))
				}
			} else {
				for s := 0; s < 4; s++ {
					avail = append(avail, fmt.Sprintf("U%d.M%d.S%d", u, m, s))
				}
			}
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(16, avail, ZS5)
	}
}
