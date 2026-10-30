package main

import (
	"errors"
	"fmt"
	"math"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	ErrOverflow  = errors.New("number exceeds maximum limit")
	ErrUnderflow = errors.New("number falls below minimum limit")
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

func generateId(prefix Prefix, num int) (ID, error) {

	max := 7
	min := 3

	if num < min || num > max {
		return "", fmt.Errorf("Id Number must between %d and %d ", min, max)
	}

	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	generated, err := gonanoid.Generate(alphabet, num)

	if err != nil {
		return "", fmt.Errorf("cannot generate id: %w", err)
	}

	fullid := string(prefix) + generated

	return ID(fullid), nil
}


