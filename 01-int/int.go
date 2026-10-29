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

const maxUint8 = math.MaxUint8

const (
	ItemActive		Flag = 1 << iota
	ItemRestricted
	ItemExpired
)

type (
	Flag 	uint8
	IDs 	string
)
type Inventory struct {
	ID 			IDs
	Stock 		uint8
	ItemFlag 	Flag
}

func generateId() (IDs, error) {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	generated, err := gonanoid.Generate(alphabet, 5) 
	
	if err != nil {
   		return "", fmt.Errorf("cannot generate id: %w", err)
}

	return IDs(generated), nil
}
