package main

import (
	"errors"
	"fmt"
	"math"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	ErrOverflow    = errors.New("Number exceeds maximum limit")
	ErrUnderflow   = errors.New("Number falls below minimum limit")
	ErrEmptyString = errors.New("Input cannot be empty")
	ErrZeroStock   = errors.New("Out of stock")
	ErrNotFound    = errors.New("ID not found")

	SetTemplate = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	InventoryDict    = make(map[ID]Inventory, 4)
	TransactionsDict = make(map[ID]Transactions, 9)
	EmployeeDict     = make(map[ID]Employee, 2)
	CustDict         = make(map[ID]Customer, 4)
)

const (
	ItemActive Flag = 1 << iota
	ItemRestricted
	ItemExpired
)

const (
	Item  Prefix = "ITM-"
	Trans Prefix = "TRN-"
	Emp   Prefix = "EMP-"
	Cust  Prefix = "CST-"

	MaxUint8 = math.MaxUint8
	MinUint8 = 0
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

func Epoch() int64 {
	return time.Now().Unix()
}

func GetTimestamp(sec int64) string {
	time := time.Unix(sec, 0)

	return time.Format("2006-01-02 15:04:05")
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

func IsExistID[V any](dict map[ID]V, id ID) bool {
	_, exists := dict[id]
	return exists
}

func CheckId[V any](prefix Prefix, num int, dict map[ID]V) (ID, error) {
	var id ID
	var err error

	for {
		id, err = generateId(prefix, num)
		if err != nil {
			return "", err
		}

		if !IsExistID(dict, id) {
			break
		}
	}

	return id, err
}

func generateId(prefix Prefix, num int) (ID, error) {

	max := 7
	min := 3

	if num < min || num > max {
		return "", fmt.Errorf("Id Number must between %d and %d ", min, max)
	}

	generated, err := gonanoid.Generate(SetTemplate, num)

	if err != nil {
		return "", fmt.Errorf("cannot generate id: %w", err)
	}

	fullid := string(prefix) + generated

	return ID(fullid), nil
}

func (i *Inventory) Record(empID ID, custID ID, amount uint8, desc string, isAdd bool) (*Transactions, error) {
	id, err := CheckId(Item, 4, TransactionsDict)

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