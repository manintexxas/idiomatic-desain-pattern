package main

import (
		"fmt"
		"math"
		"errors"

		gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
    ErrOverflow  = errors.New("number exceeds maximum limit")
    ErrUnderflow = errors.New("number falls below minimum limit")
)

const maxInt8 = math.MaxInt8

const (
	ItemActive		Flag = 1 << iota
	ItemRestricted
	ItemExpired
)

type (
	Flag 	uint8
	IDs 	string
)

func (id *IDs) generateId() (string, error) {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	generated, err := gonanoid.Generate(alphabet, 5) 
	
	if err != nil {
   		return "", fmt.Errorf("cannot generate id: %w", err)
	}

	return generated, nil
}

type Inventory struct {
	ID 			IDs
	Stock 		int8
	ItemFlag 	Flag
}