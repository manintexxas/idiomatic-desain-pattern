package model

type ProcessResponseDTO struct {
	Mode        string              `json:"mode"`
	Valid       bool                `json:"valid"`
	Output      string              `json:"output"`
	CustomRules EffectiveRulesDTO   `json:"custom_rules"`
	Report      ViolationReportDTO  `json:"report"`
	Telemetry   ProcessTelemetryDTO `json:"telemetry"`
}

type ProcessTelemetryDTO struct {
	SetupNS     int64 `json:"setup_ns"`
	ProcessNS   int64 `json:"process_ns"`
	TotalNS     int64 `json:"total_ns"`
	InputBytes  int   `json:"input_bytes"`
	OutputBytes int   `json:"output_bytes"`
}
type ViolationReportDTO struct {
	Total      int                `json:"total"`
	Violations []ViolationItemDTO `json:"violations"`
}

type ViolationItemDTO struct {
	Cause    string `json:"cause"`
	Contains int    `json:"contains"`
}

type ErrorResponseDTO struct {
	Error string `json:"error"`
}

type PresetOptionDTO struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type EffectiveRulesDTO struct {
	MinLength     int `json:"min_length"`
	MaxLength     int `json:"max_length"`

	MaxLetter     int `json:"max_letter"`
	MaxUpper      int `json:"max_upper"`
	MaxLower      int `json:"max_lower"`
	MaxNumbers    int `json:"max_numbers"`
	MaxSymbols    int `json:"max_symbols"`
	MaxWhiteSpace int `json:"max_whitespace"`

	ReqLetter     int `json:"req_letter"`
	ReqUpper      int `json:"req_upper"`
	ReqLower      int `json:"req_lower"`
	ReqNumbers    int `json:"req_numbers"`
	ReqSymbols    int `json:"req_symbols"`
	ReqWhiteSpace int `json:"req_whitespace"`
}

type EngineConfigurationDTO struct {
	Modes                []string          `json:"modes"`
	DefaultRules         EffectiveRulesDTO `json:"default_rules"`
	DefaultCharacterFlags []string         `json:"default_character_flags"`
	MaxPayloadBytes      int               `json:"max_payload_bytes"`
}
