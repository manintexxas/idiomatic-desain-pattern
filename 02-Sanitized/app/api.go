	package app

	import "errors"

	var (
		ErrInvalidOptions            = errors.New("invalid API options")
		ErrInvalidEngineMode         = errors.New("exactly one engine mode must be selected")
		ErrMissingKeywordList        = errors.New("keyword filtering requires a keyword list")
		ErrWhitelistRequiresKeywords = errors.New("whitelist requires keyword filtering")

		ErrInvalidPreset = ErrInvalidOptions
	)

	type APIOptions struct {
		CharacterFlags       Flag
		Mode                 EngineMode
		Rules                CostumeRules
		Keywords             []string
		Whitelist            []string
		KeywordFilterEnabled bool
		MaskingByte          byte
	}

	func DefaultAPIOptions() APIOptions {
		return APIOptions{
			CharacterFlags: DefaultCharacterFlags,
			Mode:           ModeFind,
			Rules:          DefaultRules(),
		}
	}

	type API struct {
		config Config
		rules  CostumeRules
		engine *Engine
	}

	type APIResult struct {
		Output []byte
		Report ViolationReport
		Err    error
	}

	func (r APIResult) Valid() bool {
		return r.Err == nil && !r.Report.HasViolations()
	}

	func NewAPI(options APIOptions) (*API, error) {
		if err, invalid := options.Rules.ValidateRulesConfig(); invalid {
			if err != nil {
				return nil, err
			}
			return nil, ErrInvalidOptions
		}

		if options.CharacterFlags & ^DefaultCharacterFlags != 0 {
			return nil, ErrInvalidOptions
		}

		mode := options.Mode
		if mode == 0 {
			mode = ModeFind
		}

		switch mode {
		case ModeFind, ModeClean, ModeMask:
		default:
			return nil, ErrInvalidEngineMode
		}

		hasKeywords := hasNonEmptyEntries(options.Keywords)
		hasWhitelist := hasNonEmptyEntries(options.Whitelist)
		keywordEnabled := options.KeywordFilterEnabled || hasKeywords

		if keywordEnabled && !hasKeywords {
			return nil, ErrMissingKeywordList
		}
		if hasWhitelist && !hasKeywords {
			return nil, ErrWhitelistRequiresKeywords
		}

		config := Config{
			UniversalFlags: options.CharacterFlags,
			EngineModes:    mode,
			MaskingByte:    options.MaskingByte,
		}

		if keywordEnabled {
			config.UniversalFlags |= FlagAllowKeywords
		}

		if config.MaskingByte == 0 {
			config.DefaultMaskingByte()
		}

		engine := buildAPIEngine(options.Keywords, options.Whitelist)

		return &API{
			config: config,
			rules:  options.Rules,
			engine: engine,
		}, nil
	}

	func hasNonEmptyEntries(values []string) bool {
		for _, value := range values {
			if value != "" {
				return true
			}
		}
		return false
	}

	func buildAPIEngine(keywords, whitelist []string) *Engine {
		if !hasNonEmptyEntries(keywords) && !hasNonEmptyEntries(whitelist) {
			return nil
		}

		engine := &Engine{}

		if hasNonEmptyEntries(keywords) {
			engine.Trie = NewTrie()
			for _, keyword := range keywords {
				if keyword == "" {
					continue
				}
				engine.Trie.Insert([]byte(keyword))
			}
		}

		if hasNonEmptyEntries(whitelist) {
			engine.WlTrie = NewWhitelistTrie()
			for _, word := range whitelist {
				if word == "" {
					continue
				}
				engine.WlTrie.Insert([]byte(word))
			}
		}

		return engine
	}

	func (a *API) Process(input []byte) APIResult {
		if a == nil {
			return APIResult{Err: ErrValidationFailed}
		}

		var report ViolationReport
		config := a.config

		output := ProcessPipeline(
			input,
			a.rules,
			&config,
			a.engine,
			&report,
		)

		return APIResult{
			Output: output,
			Report: report,
		}
	}

	func (a *API) ProcessString(input string) APIResult {
		return a.Process([]byte(input))
	}
