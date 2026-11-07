package app

type CostumeRules struct {
	MinLength 		int
	MaxLength 		int
	MaxUpper 		int
	MaxLower 		int
	MaxNumbers 		int
	MaxSymbols 		int
	MaxWhiteSpace 	int
	ReqUpper 		int
	ReqLower 		int
	ReqNumbers 		int
	ReqSymbols 		int
}

func DefaultRules() CostumeRules {
	return CostumeRules{
		MinLength: 		0,
		MaxLength: 		0,
		MaxUpper: 		-1,
		MaxLower: 		-1,
		MaxNumbers: 	-1,
		MaxSymbols: 	-1,
		MaxWhiteSpace: 	-1,
		ReqUpper: 		0,
		ReqLower: 		0,
		ReqNumbers: 	0,
		ReqSymbols: 	0,
	}
}