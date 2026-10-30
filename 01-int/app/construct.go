package app

import (
	"time"

    "dict/model"
    "dict/utils"
)

func NewInventory(date string, itemname string, stock uint8, flag model.Flag) (*model.Inventory, error) {
	parsedTime, err := time.Parse(model.Layout, date)
	if err != nil {
		return nil, err
	}

	cleanedName, err := utils.NameFormat(itemname)
	if err != nil {
		return nil, err
	}

	id, err := utils.CheckId(model.Item, 4, model.InventoryDict)
	if err != nil {
		return nil, err
	}

	inv := model.Inventory{
		ItemExp:  parsedTime.Unix(),
		ItemID:   id,
		ItemName: cleanedName,
		Stock:    stock,
		ItemFlag: flag,
	}

	model.InventoryDict[id] = inv

	return &inv, nil
}

func NewEmployee(name string) (*model.Employee, error) {
	cleanedName, err := utils.NameFormat(name)
	if err != nil {
		return nil, err
	}

	id, err := utils.CheckId(model.Emp, 4, model.InventoryDict)
	if err != nil {
		return nil, err
	}

	emp := model.Employee{
		EmpID:   id,
		EmpName: cleanedName,
	}

	model.EmployeeDict[id] = emp

	return &emp, nil
}

func NewCustomer(name string, member bool) (*model.Customer, error) {
	cleanedName, err := utils.NameFormat(name)
	if err != nil {
		return nil, err
	}

	id, err := utils.CheckId(model.Cust, 4, model.InventoryDict)
	if err != nil {
		return nil, err
	}

	cst := model.Customer{
		CustID:   id,
		CustName: cleanedName,
		IsMember: member,
	}

	model.CustDict[id] = cst

	return &cst, nil
}
