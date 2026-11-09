package model

type CustomRulesDTO struct {
	MinLength     *int `json:"min_length,omitempty"`
	MaxLength     *int `json:"max_length,omitempty"`

	MaxLetter     *int `json:"max_letter,omitempty"`
	MaxUpper      *int `json:"max_upper,omitempty"`
	MaxLower      *int `json:"max_lower,omitempty"`
	MaxNumbers    *int `json:"max_numbers,omitempty"`
	MaxSymbols    *int `json:"max_symbols,omitempty"`
	MaxWhiteSpace *int `json:"max_whitespace,omitempty"`

	ReqLetter     *int `json:"req_letter,omitempty"`
	ReqUpper      *int `json:"req_upper,omitempty"`
	ReqLower      *int `json:"req_lower,omitempty"`
	ReqNumbers    *int `json:"req_numbers,omitempty"`
	ReqSymbols    *int `json:"req_symbols,omitempty"`
	ReqWhiteSpace *int `json:"req_whitespace,omitempty"`
}

type ProcessRequestDTO struct {
	Payload        string          `json:"payload"`
	Mode           string          `json:"mode,omitempty"`
	CharacterFlags *[]string       `json:"character_flags,omitempty"`
	Keywords       []string        `json:"keywords,omitempty"`
	Whitelist      []string        `json:"whitelist,omitempty"`
	CustomRules    *CustomRulesDTO `json:"custom_rules,omitempty"`
}
