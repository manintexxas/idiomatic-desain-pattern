package app

import (
	"time"
)

func NewInventory(date string, itemname string, stock uint8, flag Flag) (*Inventory, error) {
	parsedTime, err := time.Parse(Layout, date)
	if err != nil {
		return nil, err
	}

	cleanedName, err := NameFormat(itemname)
	if err != nil {
		return nil, err
	}

	id, err := CheckId(Item, 4, InventoryDict)
	if err != nil {
		return nil, err
	}

	inv := Inventory{
		ItemExp:  parsedTime.Unix(),
		ItemID:   id,
		ItemName: cleanedName,
		Stock:    stock,
		ItemFlag: flag,
	}

	InventoryDict[id] = inv

	return &inv, nil
}

func NewEmployee(name string) (*Employee, error) {
	cleanedName, err := NameFormat(name)
	if err != nil {
		return nil, err
	}

	id, err := CheckId(Emp, 4, EmployeeDict)
	if err != nil {
		return nil, err
	}

	emp := Employee{
		EmpID:   id,
		EmpName: cleanedName,
	}

	EmployeeDict[id] = emp

	return &emp, nil
}

func NewCustomer(name string, member bool) (*Customer, error) {
	cleanedName, err := NameFormat(name)
	if err != nil {
		return nil, err
	}

	id, err := CheckId(Cust, 4, CustDict)
	if err != nil {
		return nil, err
	}

	cst := Customer{
		CustID:   id,
		CustName: cleanedName,
		IsMember: member,
	}

	CustDict[id] = cst

	return &cst, nil
}
