package app

import (
		"fmt"
)

type (
	Flag   uint8
	ID     string
	Prefix string
)

type Employee struct {
	EmpID   ID
	EmpName string
}

type Customer struct {
	CustID   ID
	CustName string

	IsMember bool
}
type Inventory struct {
	ItemExp int64

	ItemID   ID
	ItemName string
	Stock    uint8

	ItemFlag Flag
}

type Transactions struct {
	Timestamp int64
	TransID   ID
	ItemID    ID

	EmpID  ID
	CustID ID

	Desc string

	Amount   uint8
	CurStock uint8
	NowStock uint8
}

func (i *Inventory) IsActive() bool {
	return (i.ItemFlag & ItemActive) != 0
}

func (i *Inventory) IsResticted() bool {
	return (i.ItemFlag & ItemRestricted) != 0
}

func (i *Inventory) IsExpired() bool {
	return (i.ItemFlag & ItemExpired) != 0
}

func (i *Inventory) Record(empID ID, custID ID, amount uint8, desc string, isAdd bool) (*Transactions, error) {
	id, err := CheckId(Trans, 4, 	TransactionsDict)

	if err != nil {
		return nil, err
	}

	currentStock := i.Stock

	if isAdd {
		i.Stock += amount
		custID = "_"

	} else {
		i.Stock -= amount
	}

	t := Transactions{
		Timestamp: Epoch(),
		TransID:   id,
		ItemID:    i.ItemID,
		EmpID:     empID,
		CustID:    custID,
		Desc:      desc,
		Amount:    amount,
		CurStock:  currentStock,
		NowStock:  i.Stock,
	}

	TransactionsDict[id] = t

	return &t, nil
}

func (i *Inventory) AddStock(receiver ID, amount uint8, desc string) (*Transactions, error) {

	if !IsExistID(EmployeeDict, receiver) {
		return nil, fmt.Errorf("%w: Reciever id is not valid", ErrEmptyString)
	}

	if i.Stock > MaxUint8-amount {
		return nil, fmt.Errorf("%w: Stock out of limit", ErrOverflow)
	}

	if desc == "" {
		return nil, fmt.Errorf("%w: Please write somethings", ErrEmptyString)
	}

	return i.Record(receiver, "", amount, desc, true)
}

func (i *Inventory) ReduceStock(empID ID, custID ID, amount uint8, desc string) (*Transactions, error) {
	if !IsExistID(CustDict, custID) {
		return nil, fmt.Errorf("%w: Customer id not exist", ErrNotFound)
	}

	if !IsExistID(EmployeeDict, empID) {
		return nil, fmt.Errorf("%w: Employee id not exist", ErrNotFound)
	}

	if i.Stock == 0 {
		return nil, fmt.Errorf("%w: %s stock is empty", ErrZeroStock, i.ItemName)
	}

	if i.Stock < amount {
		return nil, fmt.Errorf("%w: Amount out of limit", ErrOverflow)
	}

	if desc == "" {
		desc = "Normal Transaction"
	}
	
	if !(i.IsActive() || i.IsResticted()) {
		return nil, fmt.Errorf("%w: %s is expired", ErrZeroStock, i.ItemName)
	}

	if i.IsResticted() {
		cust := CustDict[custID]

		if !cust.IsMember {
			return nil, fmt.Errorf("%w: Item member only", ErrZeroStock)
		}
	}

	return i.Record(empID, custID, amount, desc, false)
}