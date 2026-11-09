
package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	app "02-Sanitized/app"
	model "02-Sanitized/model"
)

const (
	defaultHTTPAddress = ":8080"

	maxRequestBodyBytes int64 = 1 << 20 // 1 MiB
)

// ============================================================
// SERVER
// ============================================================

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = defaultHTTPAddress
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           newHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer stop()

	go func() {
		<-ctx.Done()

		log.Println("shutting down HTTP server")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}()

	log.Printf("GTSGE Engine listening on %s", addr)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server failed: %v", err)
	}

	log.Println("HTTP server stopped")
}

// ============================================================
// ROUTER
// ============================================================

func newHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/configuration", configurationHandler)
	mux.HandleFunc("/process", processHandler)

	// Serve wwwroot.
	mux.Handle(
		"/",
		http.FileServer(http.Dir("wwwroot")),
	)

	return securityMiddleware(
		recoveryMiddleware(mux),
	)
}

// ============================================================
// SECURITY MIDDLEWARE
// ============================================================

func securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")

		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'self'; "+
				"object-src 'none'; "+
				"base-uri 'self'; "+
				"frame-ancestors 'none'",
		)

		w.Header().Set(
			"Permissions-Policy",
			"camera=(), microphone=(), geolocation=()",
		)

		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf(
					"HTTP handler panic: %v",
					recovered,
				)

				writeError(
					w,
					http.StatusInternalServerError,
					"internal server error",
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ============================================================
// RESPONSE HELPERS
// ============================================================

func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON response encoding failed: %v", err)
	}
}

func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(
		w,
		status,
		model.ErrorResponseDTO{
			Error: message,
		},
	)
}

func requireMethod(
	w http.ResponseWriter,
	r *http.Request,
	expected string,
) bool {
	if r.Method == expected {
		return true
	}

	w.Header().Set("Allow", expected)

	writeError(
		w,
		http.StatusMethodNotAllowed,
		"method not allowed",
	)

	return false
}

// ============================================================
// REQUEST DECODER
// ============================================================

var errUnsupportedMediaType = errors.New(
	"content type must be application/json",
)

func decodeJSONBody(
	w http.ResponseWriter,
	r *http.Request,
	destination any,
) error {
	contentType := r.Header.Get("Content-Type")

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return errUnsupportedMediaType
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxRequestBodyBytes,
	)

	decoder := json.NewDecoder(r.Body)

	// Reject unknown JSON fields.
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	// Reject multiple JSON values in one request.
	var extra any

	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New(
				"request body must contain exactly one JSON value",
			)
		}

		return err
	}

	return nil
}

func handleDecodeError(
	w http.ResponseWriter,
	err error,
) {
	var maxBytesErr *http.MaxBytesError

	switch {
	case errors.As(err, &maxBytesErr):
		writeError(
			w,
			http.StatusRequestEntityTooLarge,
			"request body exceeds maximum allowed size",
		)

	case errors.Is(err, errUnsupportedMediaType):
		writeError(
			w,
			http.StatusUnsupportedMediaType,
			err.Error(),
		)

	default:
		writeError(
			w,
			http.StatusBadRequest,
			"invalid JSON request",
		)
	}
}

// ============================================================
// CONFIGURATION ENDPOINT
// ============================================================

func configurationHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	defaultRules := app.DefaultRules()

	response := model.EngineConfigurationDTO{
		Modes: []string{
			"find",
			"clean",
			"mask",
		},

		DefaultRules: toEffectiveRulesDTO(defaultRules),

		DefaultCharacterFlags: defaultCharacterFlagNames(),

		MaxPayloadBytes: app.MaxAllowedPayloadSize,
	}

	writeJSON(
		w,
		http.StatusOK,
		response,
	)
}

// ============================================================
// PROCESS ENDPOINT
// ============================================================

func processHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var request model.ProcessRequestDTO

	if err := decodeJSONBody(w, r, &request); err != nil {
		handleDecodeError(w, err)
		return
	}

	totalStarted := time.Now()
	setupStarted := time.Now()

	// --------------------------------------------------------
	// 1. ENGINE MODE
	// --------------------------------------------------------

	mode, modeName, err := parseEngineMode(request.Mode)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	// --------------------------------------------------------
	// 2. CHARACTER FLAGS
	// --------------------------------------------------------

	characterFlags, err := parseCharacterFlags(
		request.CharacterFlags,
	)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	// --------------------------------------------------------
	// 3. CUSTOM RULES MERGE
	// --------------------------------------------------------

	rules := mergeCustomRules(
		app.DefaultRules(),
		request.CustomRules,
	)

	// --------------------------------------------------------
	// 4. BUILD API OPTIONS
	// --------------------------------------------------------

	options := app.DefaultAPIOptions()

	options.Mode = mode
	options.CharacterFlags = characterFlags
	options.Rules = rules
	options.Keywords = request.Keywords
	options.Whitelist = request.Whitelist

	// MaskingByte intentionally remains zero.
	// app.NewAPI() will call DefaultMaskingByte().

	// --------------------------------------------------------
	// 5. BUILD ENGINE
	// --------------------------------------------------------

	engineAPI, err := app.NewAPI(options)

	setupNS := time.Since(setupStarted).Nanoseconds()

	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	// --------------------------------------------------------
	// 6. PROCESS INPUT
	// --------------------------------------------------------

	payload := []byte(request.Payload)

	processStarted := time.Now()

	result := engineAPI.Process(payload)

	processNS := time.Since(processStarted).Nanoseconds()

	if result.Err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			"engine processing failed",
		)
		return
	}

	// --------------------------------------------------------
	// 7. VIOLATION REPORT MAPPING
	// --------------------------------------------------------

	detail := result.Report.GetErrorsDetail()

	violations := make(
		[]model.ViolationItemDTO,
		0,
		len(detail.Violations),
	)

	for _, item := range detail.Violations {
		violations = append(
			violations,
			model.ViolationItemDTO{
				Cause:    item.Cause,
				Contains: item.Contains,
			},
		)
	}

	report := model.ViolationReportDTO{
		Total:      detail.TotalViolations,
		Violations: violations,
	}

	// --------------------------------------------------------
	// 8. TELEMETRY
	// --------------------------------------------------------

	totalNS := time.Since(totalStarted).Nanoseconds()

	telemetry := model.ProcessTelemetryDTO{
		SetupNS:     setupNS,
		ProcessNS:   processNS,
		TotalNS:     totalNS,
		InputBytes:  len(payload),
		OutputBytes: len(result.Output),
	}

	// --------------------------------------------------------
	// 9. PROCESS RESPONSE
	// --------------------------------------------------------

	response := model.ProcessResponseDTO{
		Mode:        modeName,
		Valid:       result.Valid(),
		Output:      string(result.Output),
		CustomRules: toEffectiveRulesDTO(rules),
		Report:      report,
		Telemetry:   telemetry,
	}

	writeJSON(
		w,
		http.StatusOK,
		response,
	)
}

// ============================================================
// ENGINE MODE MAPPING
// ============================================================

func parseEngineMode(
	value string,
) (app.EngineMode, string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "find":
		return app.ModeFind, "find", nil

	case "clean":
		return app.ModeClean, "clean", nil

	case "mask":
		return app.ModeMask, "mask", nil

	default:
		return 0, "", errors.New(
			"unsupported engine mode",
		)
	}
}

// ============================================================
// CHARACTER FLAGS MAPPING
// ============================================================

func parseCharacterFlags(
	input *[]string,
) (app.Flag, error) {
	// Omitted field means default engine flags.
	if input == nil {
		return app.DefaultCharacterFlags, nil
	}

	var flags app.Flag

	seen := make(map[string]struct{}, len(*input))

	for _, value := range *input {
		name := strings.ToLower(strings.TrimSpace(value))

		if _, exists := seen[name]; exists {
			return 0, errors.New(
				"duplicate character flag: " + name,
			)
		}

		seen[name] = struct{}{}

		switch name {
		case "uppercase":
			flags |= app.FlagAllowUpperOnly

		case "lowercase":
			flags |= app.FlagAllowLowerOnly

		case "numbers":
			flags |= app.FlagAllowNumbers

		case "whitespace":
			flags |= app.FlagAllowWhitespace

		case "symbols":
			flags |= app.FlagAllowSymbols

		default:
			return 0, errors.New(
				"unsupported character flag: " + name,
			)
		}
	}

	return flags, nil
}

func defaultCharacterFlagNames() []string {
	flags := app.DefaultCharacterFlags

	names := make([]string, 0, 5)

	if flags&app.FlagAllowUpperOnly != 0 {
		names = append(names, "uppercase")
	}

	if flags&app.FlagAllowLowerOnly != 0 {
		names = append(names, "lowercase")
	}

	if flags&app.FlagAllowNumbers != 0 {
		names = append(names, "numbers")
	}

	if flags&app.FlagAllowWhitespace != 0 {
		names = append(names, "whitespace")
	}

	if flags&app.FlagAllowSymbols != 0 {
		names = append(names, "symbols")
	}

	return names
}

// ============================================================
// CUSTOM RULES MERGE
// ============================================================

func mergeCustomRules(
	base app.CostumeRules,
	override *model.CustomRulesDTO,
) app.CostumeRules {
	if override == nil {
		return base
	}

	if override.MinLength != nil {
		base.MinLength = *override.MinLength
	}

	if override.MaxLength != nil {
		base.MaxLength = *override.MaxLength
	}

	if override.MaxLetter != nil {
		base.MaxLetter = *override.MaxLetter
	}

	if override.MaxUpper != nil {
		base.MaxUpper = *override.MaxUpper
	}

	if override.MaxLower != nil {
		base.MaxLower = *override.MaxLower
	}

	if override.MaxNumbers != nil {
		base.MaxNumbers = *override.MaxNumbers
	}

	if override.MaxSymbols != nil {
		base.MaxSymbols = *override.MaxSymbols
	}

	if override.MaxWhiteSpace != nil {
		base.MaxWhiteSpace = *override.MaxWhiteSpace
	}

	if override.ReqLetter != nil {
		base.ReqLetter = *override.ReqLetter
	}

	if override.ReqUpper != nil {
		base.ReqUpper = *override.ReqUpper
	}

	if override.ReqLower != nil {
		base.ReqLower = *override.ReqLower
	}

	if override.ReqNumbers != nil {
		base.ReqNumbers = *override.ReqNumbers
	}

	if override.ReqSymbols != nil {
		base.ReqSymbols = *override.ReqSymbols
	}

	if override.ReqWhiteSpace != nil {
		base.ReqWhiteSpace = *override.ReqWhiteSpace
	}

	return base
}

// ============================================================
// EFFECTIVE RULES RESPONSE MAPPING
// ============================================================

func toEffectiveRulesDTO(
	rules app.CostumeRules,
) model.EffectiveRulesDTO {
	return model.EffectiveRulesDTO{
		MinLength:     rules.MinLength,
		MaxLength:     rules.MaxLength,

		MaxLetter:     rules.MaxLetter,
		MaxUpper:      rules.MaxUpper,
		MaxLower:      rules.MaxLower,
		MaxNumbers:    rules.MaxNumbers,
		MaxSymbols:    rules.MaxSymbols,
		MaxWhiteSpace: rules.MaxWhiteSpace,

		ReqLetter:     rules.ReqLetter,
		ReqUpper:      rules.ReqUpper,
		ReqLower:      rules.ReqLower,
		ReqNumbers:    rules.ReqNumbers,
		ReqSymbols:    rules.ReqSymbols,
		ReqWhiteSpace: rules.ReqWhiteSpace,
	}
}
