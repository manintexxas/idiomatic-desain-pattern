package app

import "testing"

func BenchmarkNewInventory(b *testing.B) {
	resetState()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = NewInventory("2026-12-31 00:00:00", "Keyboard Mech", 50, ItemActive)
	}
}

func BenchmarkNewEmployee(b *testing.B) {
	resetState()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = NewEmployee("Ahmad Dahlan")
	}
}

func BenchmarkNewCustomer(b *testing.B) {
	resetState()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = NewCustomer("Dewi Lestari", true)
	}
}
