package palladiumalloc

import (
	"fmt"
	"testing"
)

// 生成指定LD数量的完整可用Domain列表
func generateFullAvailable(ldCount int) []string {
	avail := make([]string, 0, ldCount*domainsPerLD)
	for ld := 0; ld < ldCount; ld++ {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	return avail
}

func BenchmarkAllocate_1Domain(b *testing.B) {
	avail := generateFullAvailable(6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(1, avail, PZ1)
	}
}

func BenchmarkAllocate_4Domains(b *testing.B) {
	avail := generateFullAvailable(6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(4, avail, PZ1)
	}
}

func BenchmarkAllocate_8Domains(b *testing.B) {
	avail := generateFullAvailable(6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(8, avail, PZ1)
	}
}

func BenchmarkAllocate_9Domains(b *testing.B) {
	avail := generateFullAvailable(12)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(9, avail, PZ1)
	}
}

func BenchmarkAllocate_48Domains(b *testing.B) {
	avail := generateFullAvailable(6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(48, avail, PZ1)
	}
}

func BenchmarkAllocate_49Domains(b *testing.B) {
	avail := generateFullAvailable(18)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(49, avail, PZ1)
	}
}

func BenchmarkAllocate_Scale_6LDs(b *testing.B)   { benchScale(b, 6) }
func BenchmarkAllocate_Scale_18LDs(b *testing.B)  { benchScale(b, 18) }
func BenchmarkAllocate_Scale_72LDs(b *testing.B)  { benchScale(b, 72) }
func BenchmarkAllocate_Scale_288LDs(b *testing.B) { benchScale(b, 288) }

func benchScale(b *testing.B, ldCount int) {
	avail := generateFullAvailable(ldCount)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(1, avail, PZ1)
	}
}

func BenchmarkBuildLDMap_6LDs(b *testing.B) {
	avail := generateFullAvailable(6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildLDMap(avail)
	}
}

func BenchmarkBuildLDMap_18LDs(b *testing.B) {
	avail := generateFullAvailable(18)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildLDMap(avail)
	}
}

func BenchmarkAllocate_SparseAvailable(b *testing.B) {
	var avail []string
	for ld := 0; ld < 72; ld += 2 {
		for d := 0; d < domainsPerLD; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(8, avail, PZ1)
	}
}

func BenchmarkAllocate_PartialLDs(b *testing.B) {
	var avail []string
	for ld := 0; ld < 36; ld++ {
		for d := 0; d < 5; d++ {
			avail = append(avail, fmt.Sprintf("%d.%d", ld, d))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(4, avail, PZ1)
	}
}

func BenchmarkAllocate_LargeRequest(b *testing.B) {
	avail := generateFullAvailable(72)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Allocate(72*8, avail, PZ1)
	}
}
