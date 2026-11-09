package app

import "sync"

func init() {
	for i := range charTable {
		b := byte(i)

		switch {
		case b >= '0' && b <= '9':
			charTable[b] = FlagAllowNumbers
		case b >= 'A' && b <= 'Z':
			charTable[b] = FlagAllowUpperOnly
		case b >= 'a' && b <= 'z':
			charTable[b] = FlagAllowLowerOnly
		case (b >= 33 && b <= 47),
			(b >= 58 && b <= 64),
			(b >= 91 && b <= 96),
			(b >= 123 && b <= 126):
			charTable[b] = FlagAllowSymbols
		case (b >= 9 && b <= 13), b == 32:
			charTable[b] = FlagAllowWhitespace
		default:
			charTable[b] = 0
		}
	}
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, MaxAllowedPayloadSize)
	},
}

func GetBuffer(src []byte) []byte {
	buf := bufferPool.Get().([]byte)
	return append(buf[:0], src...)
}

func PutBuffer(buf []byte) {
	if buf == nil {
		return
	}

	full := buf[:cap(buf)]
	for i := range full {
		full[i] = 0
	}

	if cap(buf) < MaxAllowedPayloadSize || cap(buf) > MaxAllowedPayloadSize*2 {
		return
	}

	bufferPool.Put(full[:0])
}

func ProcessPipeline(
	input []byte,
	rules CostumeRules,
	config *Config,
	engine *Engine,
	report *ViolationReport,
) []byte {
	if report == nil {
		report = &ViolationReport{}
	} else {
		*report = ViolationReport{}
	}

	if config == nil {
		report.TotalViolations = 1
		return []byte{}
	}

	length := len(input)
	if length > MaxAllowedPayloadSize {
		report.TotalViolations = 1
		report.MaxLengthViolations = length - MaxAllowedPayloadSize
		return []byte{}
	}

	workBuffer := GetBuffer(input)
	defer PutBuffer(workBuffer)

	ValidateRules(workBuffer, rules, config, engine, report)

	if report.HasViolations() && report.MalformedUnicodeViolations > 0 {
		output := make([]byte, length)
		copy(output, input)
		return output
	}

	var finalLen int
	switch config.EngineModes {
	case ModeClean:
		finalLen = processCleanState(workBuffer, config, engine)
	case ModeMask:
		finalLen = processMaskState(workBuffer, config, engine)
	default:
		finalLen = processFindState(workBuffer, config, engine)
	}

	if finalLen < 0 {
		finalLen = 0
	}
	if finalLen > len(workBuffer) {
		finalLen = len(workBuffer)
	}

	output := make([]byte, finalLen)
	copy(output, workBuffer[:finalLen])
	return output
}

func processFindState(buf []byte, _ *Config, _ *Engine) int {
	return len(buf)
}

func processCleanState(buf []byte, config *Config, engine *Engine) int {
	current := buf

	if keywordFilteringEnabled(config, engine) {
		if engine.WlTrie == nil {
			current = engine.Trie.SanitizeOrMask(current, 0, true)
		} else {
			current = engine.Trie.SanitizeOrMaskWithWhitelist(current, 0, true, engine.WlTrie)
		}
	}

	writeIdx := 0
	for readIdx := 0; readIdx < len(current); readIdx++ {
		b := current[readIdx]
		if isByteViolating(b, config) {
			continue
		}
		buf[writeIdx] = b
		writeIdx++
	}

	return writeIdx
}

func processMaskState(buf []byte, config *Config, engine *Engine) int {
	maskingByte := config.MaskingByte
	if maskingByte == 0 {
		localConfig := *config
		localConfig.DefaultMaskingByte()
		maskingByte = localConfig.MaskingByte
	}

	bufLen := len(buf)

	if keywordFilteringEnabled(config, engine) {
		if engine.WlTrie == nil {
			buf = engine.Trie.SanitizeOrMask(buf[:bufLen], maskingByte, false)
		} else {
			buf = engine.Trie.SanitizeOrMaskWithWhitelist(buf[:bufLen], maskingByte, false, engine.WlTrie)
		}
		bufLen = len(buf)
	}

	for i := 0; i < bufLen; i++ {
		if isByteViolating(buf[i], config) {
			buf[i] = maskingByte
		}
	}

	return bufLen
}

func keywordFilteringEnabled(config *Config, engine *Engine) bool {
	return config != nil &&
		config.UniversalFlags&FlagAllowKeywords != 0 &&
		engine != nil &&
		engine.Trie != nil
}

func isByteViolating(b byte, config *Config) bool {
	if config == nil {
		return true
	}

	category := charTable[b]
	if category == 0 {
		return true
	}

	return category&config.UniversalFlags == 0
}
