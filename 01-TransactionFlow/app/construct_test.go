package app

import (
	"errors"
	"testing"
)

func TestNewInventory(t *testing.T) {
	resetState()

	inv, err := NewInventory("2026-12-31 00:00:00", "Laptop Gaming", 10, ItemActive)
	if err != nil {
		t.Fatalf("NewInventory failed: %v", err)
	}

	if inv.ItemName != "Laptop Gaming" || inv.Stock != 10 || inv.ItemFlag != ItemActive {
		t.Errorf("NewInventory field mismatch: %+v", inv)
	}

	if !IsExistID(InventoryDict, inv.ItemID) {
		t.Errorf("NewInventory ID %s not saved to InventoryDict", inv.ItemID)
	}

	_, errInvalidDate := NewInventory("31-12-2026 00:00:00", "Mouse", 5, ItemActive)
	if errInvalidDate == nil {
		t.Error("NewInventory should fail on invalid date format")
	}

	_, errInvalidName := NewInventory("2026-12-31 00:00:00", "", 5, ItemActive)
	if errInvalidName == nil || !errors.Is(errInvalidName, ErrInvalidInput) {
		t.Errorf("NewInventory should fail on invalid name, got: %v", errInvalidName)
	}
}

func TestNewEmployee(t *testing.T) {
	resetState()

	emp, err := NewEmployee("Budi Santoso")
	if err != nil {
		t.Fatalf("NewEmployee failed: %v", err)
	}

	if emp.EmpName != "Budi Santoso" {
		t.Errorf("NewEmployee name mismatch: %s", emp.EmpName)
	}

	if !IsExistID(EmployeeDict, emp.EmpID) {
		t.Errorf("NewEmployee ID %s not saved to EmployeeDict", emp.EmpID)
	}

	_, errEmptyName := NewEmployee("   ")
	if errEmptyName == nil || !errors.Is(errEmptyName, ErrInvalidInput) {
		t.Errorf("NewEmployee should fail on empty name, got: %v", errEmptyName)
	}
}

func TestNewCustomer(t *testing.T) {
	resetState()

	cust, err := NewCustomer("Siti Aminah", true)
	if err != nil {
		t.Fatalf("NewCustomer failed: %v", err)
	}

	if cust.CustName != "Siti Aminah" || !cust.IsMember {
		t.Errorf("NewCustomer data mismatch: %+v", cust)
	}

	if !IsExistID(CustDict, cust.CustID) {
		t.Errorf("NewCustomer ID %s not saved to CustDict", cust.CustID)
	}

	_, errSymbol := NewCustomer("Siti 💻", false)
	if errSymbol == nil || !errors.Is(errSymbol, ErrInvalidInput) {
		t.Errorf("NewCustomer should fail on invalid symbol, got: %v", errSymbol)
	}
}
