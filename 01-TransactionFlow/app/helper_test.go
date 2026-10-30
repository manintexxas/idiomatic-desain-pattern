package app

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestEpoch(t *testing.T) {
	before := time.Now().Unix()
	e := Epoch()
	after := time.Now().Unix()

	if e < before || e > after {
		t.Errorf("Epoch() = %d; want value between %d and %d", e, before, after)
	}
}

func TestGetTimestamp(t *testing.T) {
	sec := int64(1700000000)
	got := GetTimestamp(sec)
	want := time.Unix(sec, 0).Format("2006-01-02 15:04:05")

	if got != want {
		t.Errorf("GetTimestamp(%d) = %s; want %s", sec, got, want)
	}

	gotZero := GetTimestamp(0)
	if gotZero == "" {
		t.Error("GetTimestamp(0) should return valid formatted time string")
	}
}

func TestIsExistID(t *testing.T) {
	dict := map[ID]string{
		"INV-12345": "Item A",
	}

	if !IsExistID(dict, "INV-12345") {
		t.Error("IsExistID should return true for existing key")
	}

	if IsExistID(dict, "INV-99999") {
		t.Error("IsExistID should return false for non-existing key")
	}

	var nilDict map[ID]string
	if IsExistID(nilDict, "INV-12345") {
		t.Error("IsExistID should return false for nil map")
	}
}

func TestGenerateId_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		prefix  Prefix
		num     int
		wantErr bool
	}{
		{"Valid Min Length", "INV-", MinIdLen, false},
		{"Valid Max Length", "INV-", MaxIdLen, false},
		{"Below Min Length", "INV-", MinIdLen - 1, true},
		{"Above Max Length", "INV-", MaxIdLen + 1, true},
		{"Zero Length", "INV-", 0, true},
		{"Negative Length", "INV-", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := generateId(tt.prefix, tt.num)
			if (err != nil) != tt.wantErr {
				t.Fatalf("generateId() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				expectedLen := len(string(tt.prefix)) + tt.num
				if len(string(id)) != expectedLen {
					t.Errorf("generateId() len = %d; want %d", len(string(id)), expectedLen)
				}
			}
		})
	}
}

func TestCheckId(t *testing.T) {
	dict := map[ID]string{
		"EMP-12345": "John",
	}

	id, err := CheckId("EMP-", 5, dict)
	if err != nil {
		t.Fatalf("CheckId failed: %v", err)
	}
	if IsExistID(dict, id) {
		t.Errorf("CheckId generated existing ID: %s", id)
	}

	_, errInvalid := CheckId("EMP-", 1, dict)
	if errInvalid == nil {
		t.Error("CheckId should fail for out of range length")
	}
}

func TestNameFormat_EdgeCases(t *testing.T) {
	validName := strings.Repeat("a", MinCharLen)
	maxValidName := strings.Repeat("a", MaxCharLen)
	tooShort := strings.Repeat("a", MinCharLen-1)
	tooLong := strings.Repeat("a", MaxCharLen+1)

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errIs   error
	}{
		{"Valid Name", "Laptop Dell", false, nil},
		{"Valid Min Bounds", validName, false, nil},
		{"Valid Max Bounds", maxValidName, false, nil},
		{"Empty String", "", true, ErrInvalidInput},
		{"Only Spaces", "   ", true, ErrInvalidInput},
		{"Too Short", tooShort, true, ErrOutofRange},
		{"Too Long", tooLong, true, ErrOutofRange},
		{"Invalid Symbol Emoji", "Laptop 💻", true, ErrInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NameFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NameFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.errIs != nil && !errors.Is(err, tt.errIs) {
				t.Errorf("NameFormat(%q) error = %v; want %v", tt.input, err, tt.errIs)
			}
			if !tt.wantErr && got != tt.input {
				t.Errorf("NameFormat(%q) = %q; want %q", tt.input, got, tt.input)
			}
		})
	}
}

func TestDeleteData(t *testing.T) {
	dict := map[ID]string{
		"CUST-001": "Alice",
	}

	err := DeleteData(dict, "CUST-001")
	if err != nil {
		t.Fatalf("DeleteData failed: %v", err)
	}
	if IsExistID(dict, "CUST-001") {
		t.Error("DeleteData did not remove the key")
	}

	errNotFound := DeleteData(dict, "CUST-999")
	if errNotFound == nil || !errors.Is(errNotFound, ErrNotFound) {
		t.Errorf("DeleteData expected ErrNotFound, got %v", errNotFound)
	}
}	