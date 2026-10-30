package app

import (
	"errors"
	"testing"
)

func resetState() {
	InventoryDict = make(map[ID]Inventory)
	EmployeeDict = make(map[ID]Employee)
	CustDict = make(map[ID]Customer)
	TransactionsDict = make(map[ID]Transactions)
}

func TestInventory_Flags(t *testing.T) {
	invActive := Inventory{ItemFlag: ItemActive}
	if !invActive.IsActive() {
		t.Error("Expected IsActive to be true")
	}
	if invActive.IsResticted() {
		t.Error("Expected IsRestricted to be false")
	}
	if invActive.IsExpired() {
		t.Error("Expected IsExpired to be false")
	}

	invRestricted := Inventory{ItemFlag: ItemRestricted}
	if !invRestricted.IsResticted() {
		t.Error("Expected IsResticted to be true")
	}

	invExpired := Inventory{ItemFlag: ItemExpired}
	if !invExpired.IsExpired() {
		t.Error("Expected IsExpired to be true")
	}
}

func TestInventory_AddStock(t *testing.T) {
	resetState()

	empID := ID("EMP-001")
	EmployeeDict[empID] = Employee{EmpID: empID, EmpName: "Budi"}

	inv := &Inventory{
		ItemID:   ID("INV-001"),
		ItemName: "Laptop",
		Stock:    10,
		ItemFlag: ItemActive,
	}
	InventoryDict[inv.ItemID] = *inv

	trx, err := inv.AddStock(empID, 5, "Restock gudang")
	if err != nil {
		t.Fatalf("AddStock failed: %v", err)
	}

	if inv.Stock != 15 {
		t.Errorf("Stock = %d; want 15", inv.Stock)
	}

	if trx.Amount != 5 || trx.CurStock != 10 || trx.NowStock != 15 {
		t.Errorf("Transaction details mismatch: %+v", trx)
	}

	_, errInvalidEmp := inv.AddStock("EMP-999", 5, "Restock")
	if errInvalidEmp == nil {
		t.Error("AddStock should fail with invalid employee ID")
	}

	_, errEmptyDesc := inv.AddStock(empID, 5, "")
	if errEmptyDesc == nil {
		t.Error("AddStock should fail with empty description")
	}

	_, errOverflow := inv.AddStock(empID, 250, "Overflow")
	if errOverflow == nil || !errors.Is(errOverflow, ErrOverflow) {
		t.Errorf("AddStock should fail with ErrOverflow, got %v", errOverflow)
	}
}

func TestInventory_ReduceStock(t *testing.T) {
	resetState()

	empID := ID("EMP-001")
	custID := ID("CUST-001")
	EmployeeDict[empID] = Employee{EmpID: empID, EmpName: "Budi"}
	CustDict[custID] = Customer{CustID: custID, CustName: "Ani", IsMember: false}

	inv := &Inventory{
		ItemID:   ID("INV-001"),
		ItemName: "Buku",
		Stock:    10,
		ItemFlag: ItemActive,
	}

	trx, err := inv.ReduceStock(empID, custID, 3, "Penjualan")
	if err != nil {
		t.Fatalf("ReduceStock failed: %v", err)
	}

	if inv.Stock != 7 {
		t.Errorf("Stock = %d; want 7", inv.Stock)
	}

	if trx.Amount != 3 || trx.CurStock != 10 || trx.NowStock != 7 {
		t.Errorf("Transaction details mismatch: %+v", trx)
	}

	invZero := &Inventory{ItemID: ID("INV-002"), ItemName: "Kosong", Stock: 0, ItemFlag: ItemActive}
	_, errZero := invZero.ReduceStock(empID, custID, 1, "Beli")
	if errZero == nil || !errors.Is(errZero, ErrZeroStock) {
		t.Error("ReduceStock should fail on zero stock")
	}

	invExpired := &Inventory{ItemID: ID("INV-003"), ItemName: "Expired", Stock: 10, ItemFlag: ItemExpired}
	_, errExp := invExpired.ReduceStock(empID, custID, 1, "Beli")
	if errExp == nil {
		t.Error("ReduceStock should fail on expired item")
	}

	restrictedCustID := ID("CUST-002")
	CustDict[restrictedCustID] = Customer{CustID: restrictedCustID, CustName: "Joko", IsMember: false}
	invRestricted := &Inventory{ItemID: ID("INV-004"), ItemName: "Eksklusif", Stock: 10, ItemFlag: ItemRestricted}
	_, errRestricted := invRestricted.ReduceStock(empID, restrictedCustID, 1, "Beli")
	if errRestricted == nil {
		t.Error("ReduceStock should fail for non-member on restricted item")
	}

	memberCustID := ID("CUST-003")
	CustDict[memberCustID] = Customer{CustID: memberCustID, CustName: "Siti", IsMember: true}
	_, errMember := invRestricted.ReduceStock(empID, memberCustID, 1, "Beli")
	if errMember != nil {
		t.Errorf("ReduceStock should pass for member on restricted item: %v", errMember)
	}
}
