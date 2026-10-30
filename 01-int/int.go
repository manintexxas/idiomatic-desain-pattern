package main

import (
	"errors"
	"fmt"
	"math"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	ErrOverflow  	 = errors.New("Number exceeds maximum limit")
	ErrUnderflow 	 = errors.New("Number falls below minimum limit")
	ErrEmptyString 	 = errors.New("Input cannot be empty")

	SetTemplate 	 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	InventoryDict 	 = make(map[ID]Inventory, 4)
	TransactionsDict = make(map[ID]Transactions, 9)
	EmployeeDict 	 = make(map[ID]Employee, 2)
	CustDict 	  	 = make(map[ID]Customer, 4)
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

	MaxUint8 	 = math.MaxUint8
	MinUint8	 = 0
)

type (
	Flag   uint8
	ID     string
	Prefix string
)

type Employee struct {
	EmpID	ID
	EmpName	string
}

type Customer struct {
	CustID	 ID
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

func TimeFormat(sec int64) string {
	time := time.Unix(sec, 0)

	return time.Format("2006-01-02 15:04:05")
}

func IsExist[V any](dict map[ID]V, id ID) bool {
	_, exists := dict[id]
	return exists
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

func (i *Inventory) AddStock (num uint8, desc string) (*Transactions, error) {

		if num < MinUint8 {
			return nil, fmt.Errorf("%w: Amount cannot be negative number", ErrUnderflow)
		}

		if num > MaxUint8 {
			return nil, fmt.Errorf("%w: Amount out of limit ( 255 )", ErrOverflow)
		}

		if desc == "" {
			return nil, fmt.Errorf("%w: Please write somethings", ErrEmptyString)
		}
		
		if i.Stock > MaxUint8 - num {
			return nil, fmt.Errorf("%w: Stock out of limit", ErrOverflow)
		}

		curStock := i.Stock
		i.Stock += num

		var id ID
		var err error
		
		for {
			id, err = generateId(Trans, 4)
			if err != nil {
				return nil, err
			}

			if !IsExist(TransactionsDict, id) {
				break
			}
		}

		t := Transactions{
			Timestamp: Epoch(),
			TransID: 	id,
			ItemID: 	i.ItemID,
			EmpID: 		"-",
			CustID: 	"-",
			Desc: 		desc,
			Amount: 	num,
			CurStock: 	curStock,
			NowStock: 	i.Stock,
		}

		TransactionsDict[id] = t

		return &t, nil
}


func main() {
	tran := Transactions{
		Timestamp: Epoch(),
	}

	times := time.Unix(tran.Timestamp, 0)

	formattedDate := times.Format("2006-01-02 15:04:05")
	
	fmt.Printf("Format Tanggal: %s\n", formattedDate)
}