package app

import "testing"

func BenchmarkInventory_AddStock(b *testing.B) {
	resetState()
	empID := ID("EMP-001")
	EmployeeDict[empID] = Employee{EmpID: empID, EmpName: "Budi"}

	inv := &Inventory{
		ItemID:   ID("INV-001"),
		ItemName: "Benchmark Item",
		Stock:    10,
		ItemFlag: ItemActive,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		inv.Stock = 10
		_, _ = inv.AddStock(empID, 1, "Bench Add")
	}
}

func BenchmarkInventory_ReduceStock(b *testing.B) {
	resetState()
	empID := ID("EMP-001")
	custID := ID("CUST-001")
	EmployeeDict[empID] = Employee{EmpID: empID, EmpName: "Budi"}
	CustDict[custID] = Customer{CustID: custID, CustName: "Ani", IsMember: true}

	inv := &Inventory{
		ItemID:   ID("INV-001"),
		ItemName: "Benchmark Item",
		Stock:    255,
		ItemFlag: ItemActive,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		inv.Stock = 255
		_, _ = inv.ReduceStock(empID, custID, 1, "Bench Reduce")
	}
}
