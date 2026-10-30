package app

import (
	"errors"
	"math"
	"regexp"
)

var (
	ErrOverflow     = errors.New("Number exceeds maximum limit")
	ErrUnderflow    = errors.New("Number falls below minimum limit")
	ErrEmptyString  = errors.New("Input cannot be empty")
	ErrInvalidInput = errors.New("Input is invalid")
	ErrOutofRange   = errors.New("Input out of range")
	ErrZeroStock    = errors.New("Out of stock")
	ErrNotFound     = errors.New("ID not found")

	SetTemplate   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	Layout        = "2006-01-02 15:04:05"
	SafeNameRegex = regexp.MustCompile(`^[\p{L}\s'\-]+$`)

	MinCharLen = 2
	MaxCharLen = 50

	MinIdLen = 3
	MaxIdLen = 7

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
