package app

import (
	"fmt"
	"strings"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"

)

func Epoch() int64 {
	return time.Now().Unix()
}

func GetTimestamp(sec int64) string {
	time := time.Unix(sec, 0)

	return time.Format("2006-01-02 15:04:05")
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

	if num < MinIdLen || num > MaxIdLen {
		return "", fmt.Errorf("%w: Id Number must between %d and %d ", ErrOutofRange, MinIdLen, MaxIdLen)
	}

	generated, err := gonanoid.Generate(SetTemplate, num)

	if err != nil {
		return "", fmt.Errorf("cannot generate id: %w", err)
	}

	fullid := string(prefix) + generated

	return ID(fullid), nil
}

func NameFormat(name string) (string, error) {
	trimmed := strings.TrimSpace(name)

	if trimmed == "" {
		return "", fmt.Errorf("%w: Name cannot be empty", ErrInvalidInput)
	}

	if !SafeNameRegex.MatchString(name) {
		return "", fmt.Errorf("%w: name contains invalid symbols or emojis", ErrInvalidInput)
	}

	if len(name) > MaxCharLen || len(name) < MinCharLen {
		return "", fmt.Errorf("%w: Item name must be between %d and %d character ", ErrOutofRange, MinCharLen, MaxCharLen)
	}

	return name, nil

}

func DeleteData[V any](dict map[ID]V, id ID) error {
	if !IsExistID(dict, id) {
		return fmt.Errorf("%w: ID %s tidak ditemukan", ErrNotFound, id)
	}

	delete(dict, id)
	return nil
}
