package app

import (
	"fmt"
	"testing"
)

func BenchmarkEpoch(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Epoch()
	}
}

func BenchmarkGetTimestamp(b *testing.B) {
	sec := int64(1700000000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetTimestamp(sec)
	}
}

func BenchmarkIsExistID(b *testing.B) {
	dict := make(map[ID]int, 1000)
	for i := 0; i < 1000; i++ {
		dict[ID(fmt.Sprintf("ID-%d", i))] = i
	}
	target := ID("ID-500")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsExistID(dict, target)
	}
}

func BenchmarkGenerateId(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = generateId("INV-", 8)
	}
}

func BenchmarkCheckId(b *testing.B) {
	dict := make(map[ID]bool, 1000)
	for i := 0; i < 1000; i++ {
		dict[ID(fmt.Sprintf("TEST-%d", i))] = true
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = CheckId("TEST-", 8, dict)
	}
}

func BenchmarkNameFormat_Valid(b *testing.B) {
	name := "Valid Inventory Name"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = NameFormat(name)
	}
}

func BenchmarkNameFormat_InvalidRegex(b *testing.B) {
	name := "Invalid 💻 Name!"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = NameFormat(name)
	}
}

func BenchmarkDeleteData(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		dict := map[ID]int{"KEY-1": 100}
		b.StartTimer()

		_ = DeleteData(dict, "KEY-1")
	}
}
