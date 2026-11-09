(() => {
  "use strict";

  const MAX_PAYLOAD_FALLBACK = 2048;
  const INPUT_DEBOUNCE_MS = 300;
  const MODE_DESCRIPTIONS = {
    find: "Validate input without modifying its contents.",
    clean: "Remove characters and keywords that violate the active rules.",
    mask: "Replace characters and keywords that violate the active rules."
  };

  const RULE_FIELDS = [
    "min_length",
    "max_length",
    "max_letter",
    "max_upper",
    "max_lower",
    "max_numbers",
    "max_symbols",
    "max_whitespace",
    "req_letter",
    "req_upper",
    "req_lower",
    "req_numbers",
    "req_symbols",
    "req_whitespace"
  ];

  const FLAG_NAMES = [
    "uppercase",
    "lowercase",
    "numbers",
    "whitespace",
    "symbols"
  ];

  const state = {
    configuration: null,
    mode: "find",
    characterFlags: [],
    customRules: {},
    keywords: [],
    whitelist: [],
    processing: false,
    maxPayloadBytes: MAX_PAYLOAD_FALLBACK,
    debounceTimer: null,
    requestController: null,
    requestVersion: 0,
    dialogBaseline: null
  };

  const elements = {
    systemStatus: document.getElementById("system-status"),
    systemStatusDot: document.getElementById("system-status-dot"),
    payloadInput: document.getElementById("payload-input"),
    inputByteCount: document.getElementById("input-byte-count"),
    payloadWarning: document.getElementById("payload-warning"),
    openConfiguration: document.getElementById("open-configuration"),
    resetConfiguration: document.getElementById("reset-configuration"),
    configurationSummary: document.getElementById("configuration-summary"),
    outputText: document.getElementById("output-text"),
    outputByteCount: document.getElementById("output-byte-count"),
    outputState: document.getElementById("output-state"),
    outputModeLabel: document.getElementById("output-mode-label"),
    validationStatus: document.getElementById("validation-status"),
    violationTotal: document.getElementById("violation-total"),
    setupNS: document.getElementById("setup-ns"),
    processNS: document.getElementById("process-ns"),
    totalNS: document.getElementById("total-ns"),
    telemetryInputBytes: document.getElementById("telemetry-input-bytes"),
    telemetryOutputBytes: document.getElementById("telemetry-output-bytes"),
    violationList: document.getElementById("violation-list"),
    reportState: document.getElementById("report-state"),
    requestError: document.getElementById("request-error"),
    requestErrorMessage: document.getElementById("request-error-message"),
    retryRequest: document.getElementById("retry-request"),
    engineLimitLabel: document.getElementById("engine-limit-label"),
    dialog: document.getElementById("configuration-dialog"),
    configurationForm: document.getElementById("configuration-form"),
    closeConfiguration: document.getElementById("close-configuration"),
    cancelConfiguration: document.getElementById("cancel-configuration"),
    configurationError: document.getElementById("configuration-error"),
    keywordsList: document.getElementById("keywords-list"),
    whitelistList: document.getElementById("whitelist-list"),
    addKeyword: document.getElementById("add-keyword"),
    addWhitelist: document.getElementById("add-whitelist"),
    maskingCharacter: document.getElementById("masking-character"),
    characterFlagsSummary: document.getElementById("character-flags-summary"),
    rulesSummary: document.getElementById("rules-summary"),
    keywordSummary: document.getElementById("keyword-summary"),
    letterRuleStatus: document.getElementById("letter-rule-status"),
    modeDescription: document.getElementById("mode-description"),
    discardConfirmation: document.getElementById("discard-confirmation"),
    keepEditing: document.getElementById("keep-editing"),
    discardChanges: document.getElementById("discard-changes")
  };

  const numberFormatter = new Intl.NumberFormat("en-US");
  const encoder = new TextEncoder();

  class APIError extends Error {
    constructor(message, status) {
      super(message);
      this.name = "APIError";
      this.status = status;
    }
  }

  async function requestJSON(url, options = {}) {
    let response;

    try {
      response = await fetch(url, {
        ...options,
        headers: {
          Accept: "application/json",
          ...(options.body ? { "Content-Type": "application/json" } : {}),
          ...options.headers
        }
      });
    } catch (error) {
      if (error && error.name === "AbortError") throw error;
      throw new APIError("Unable to reach the GTSGE server.", 0);
    }

    const contentType = response.headers.get("content-type") || "";
    let body = null;

    if (contentType.includes("application/json")) {
      try {
        body = await response.json();
      } catch (error) {
        throw new APIError("The server returned invalid JSON.", response.status);
      }
    } else if (!response.ok) {
      throw new APIError(`The server returned HTTP ${response.status}.`, response.status);
    }

    if (!response.ok) {
      const message = body && typeof body.error === "string"
        ? body.error
        : `Request failed with HTTP ${response.status}.`;
      throw new APIError(message, response.status);
    }

    if (body === null || typeof body !== "object") {
      throw new APIError("The server returned an unexpected response.", response.status);
    }

    return body;
  }

  function setSystemStatus(label, status = "neutral") {
    elements.systemStatus.textContent = label;
    elements.systemStatusDot.classList.toggle("is-ready", status === "ready");
    elements.systemStatusDot.classList.toggle("is-error", status === "error");
  }

  function showRequestError(message) {
    elements.requestErrorMessage.textContent = message;
    elements.retryRequest.hidden = false;
    elements.requestError.hidden = false;
  }

  function clearRequestError() {
    elements.requestErrorMessage.textContent = "";
    elements.retryRequest.hidden = true;
    elements.requestError.hidden = true;
  }

  function formatInteger(value) {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? numberFormatter.format(parsed) : "—";
  }

  function getPayloadBytes() {
    return encoder.encode(elements.payloadInput.value).length;
  }

  function updatePayloadCounter() {
    const byteCount = getPayloadBytes();
    const exceeded = byteCount > state.maxPayloadBytes;

    elements.inputByteCount.textContent = `${formatInteger(byteCount)} / ${formatInteger(state.maxPayloadBytes)} bytes`;
    elements.payloadWarning.hidden = !exceeded;
    return byteCount;
  }

  function updateControlState() {
    const disabled = !state.configuration || state.processing;
    elements.openConfiguration.disabled = disabled;
    elements.resetConfiguration.disabled = disabled;

    document.querySelectorAll("[data-mode]").forEach((button) => {
      button.disabled = state.processing;
    });
  }

  function renderMode() {
    document.querySelectorAll("[data-mode]").forEach((button) => {
      const selected = button.dataset.mode === state.mode;
      button.classList.toggle("is-selected", selected);
      button.setAttribute("aria-pressed", String(selected));
    });

    elements.modeDescription.textContent = MODE_DESCRIPTIONS[state.mode] || MODE_DESCRIPTIONS.find;
    elements.outputModeLabel.textContent = state.mode.toUpperCase();
  }

  function renderConfigurationSectionSummaries() {
    const flags = readCharacterFlagsDraft();
    if (elements.characterFlagsSummary) {
      elements.characterFlagsSummary.textContent = `${flags.length} of 5 enabled`;
    }

    const ruleCount = Object.keys(state.customRules).length;
    if (elements.rulesSummary) {
      elements.rulesSummary.textContent = ruleCount ? `${ruleCount} overrides` : "No overrides";
    }

    if (elements.keywordSummary) {
      const blocked = getDraftListValues(elements.keywordsList).length;
      const allowed = getDraftListValues(elements.whitelistList).length;
      elements.keywordSummary.textContent = blocked || allowed
        ? `${blocked} blocked · ${allowed} allowed`
        : "No keywords";
    }
  }

  function renderConfigurationSummary() {
    const parts = [];
    const ruleCount = Object.keys(state.customRules).length;
    const flagCount = state.characterFlags.length;

    if (flagCount !== FLAG_NAMES.length) {
      parts.push(`${flagCount}/5 character categories`);
    }
    if (ruleCount > 0) parts.push(`${ruleCount} rule overrides`);
    if (state.keywords.length > 0) parts.push(`${state.keywords.length} keywords`);
    if (state.whitelist.length > 0) parts.push(`${state.whitelist.length} whitelist entries`);

    elements.configurationSummary.textContent = parts.length > 0
      ? parts.join(" · ")
      : "Default engine configuration";

    elements.maskingCharacter.textContent = state.characterFlags.includes("symbols") ? "*" : "x";

    if (elements.characterFlagsSummary) {
      elements.characterFlagsSummary.textContent = `${flagCount} of 5 enabled`;
    }
    if (elements.rulesSummary) {
      elements.rulesSummary.textContent = ruleCount > 0 ? `${ruleCount} overrides` : "No overrides";
    }
    if (elements.keywordSummary) {
      const blocked = state.keywords.length;
      const allowed = state.whitelist.length;
      elements.keywordSummary.textContent = blocked || allowed
        ? `${blocked} blocked · ${allowed} allowed`
        : "No keywords";
    }
  }

  function clearResult() {
    elements.outputText.value = "";
    elements.outputByteCount.textContent = "0 bytes";
    elements.outputState.textContent = "Waiting for input";
    elements.validationStatus.textContent = "Awaiting result";
    elements.validationStatus.className = "validation-status status-pending";
    elements.violationTotal.textContent = "—";
    elements.setupNS.textContent = "—";
    elements.processNS.textContent = "—";
    elements.totalNS.textContent = "—";
    elements.telemetryInputBytes.textContent = "—";
    elements.telemetryOutputBytes.textContent = "—";
    elements.reportState.textContent = "No report";
    renderEmptyReport("Violation details will appear after automatic processing.");
  }

  function renderEmptyReport(message) {
    elements.violationList.replaceChildren();
    const paragraph = document.createElement("p");
    paragraph.className = "empty-state";
    paragraph.textContent = message;
    elements.violationList.append(paragraph);
  }

  function renderViolationReport(report) {
    const violations = Array.isArray(report?.violations) ? report.violations : [];
    const total = Number.isFinite(Number(report?.total)) ? Number(report.total) : 0;

    elements.violationTotal.textContent = formatInteger(total);
    elements.reportState.textContent = `${violations.length} cause${violations.length === 1 ? "" : "s"}`;
    elements.violationList.replaceChildren();

    if (violations.length === 0) {
      renderEmptyReport("No violations reported by the engine.");
      return;
    }

    const table = document.createElement("table");
    table.className = "violation-table";

    const thead = document.createElement("thead");
    const headerRow = document.createElement("tr");
    const causeHeader = document.createElement("th");
    const countHeader = document.createElement("th");
    causeHeader.scope = "col";
    countHeader.scope = "col";
    causeHeader.textContent = "Cause";
    countHeader.textContent = "Contains";
    headerRow.append(causeHeader, countHeader);
    thead.append(headerRow);

    const tbody = document.createElement("tbody");
    for (const item of violations) {
      const row = document.createElement("tr");
      const causeCell = document.createElement("td");
      const countCell = document.createElement("td");
      causeCell.textContent = typeof item?.cause === "string" ? item.cause : "Unknown violation";
      countCell.textContent = formatInteger(item?.contains);
      row.append(causeCell, countCell);
      tbody.append(row);
    }

    table.append(thead, tbody);
    elements.violationList.append(table);
  }

  function renderProcessResponse(response) {
    if (typeof response.output !== "string" || typeof response.valid !== "boolean") {
      throw new Error("The process response is missing required output or validation fields.");
    }

    const telemetry = response.telemetry || {};
    elements.outputText.value = response.output;
    elements.outputByteCount.textContent = `${formatInteger(telemetry.output_bytes ?? encoder.encode(response.output).length)} bytes`;
    elements.outputState.textContent = "Latest response";

    elements.validationStatus.textContent = response.valid ? "Valid" : "Invalid";
    elements.validationStatus.className = `validation-status ${response.valid ? "status-valid" : "status-invalid"}`;

    elements.setupNS.textContent = formatInteger(telemetry.setup_ns);
    elements.processNS.textContent = formatInteger(telemetry.process_ns);
    elements.totalNS.textContent = formatInteger(telemetry.total_ns);
    elements.telemetryInputBytes.textContent = `${formatInteger(telemetry.input_bytes)} B`;
    elements.telemetryOutputBytes.textContent = `${formatInteger(telemetry.output_bytes)} B`;

    renderViolationReport(response.report || {});
  }

  function renderDynamicList(container, values, placeholder, removeLabel) {
    container.replaceChildren();

    if (values.length === 0) {
      const note = document.createElement("p");
      note.className = "list-empty-note";
      note.textContent = "No entries added.";
      container.append(note);
      return;
    }

    values.forEach((value, index) => {
      const row = document.createElement("div");
      row.className = "dynamic-entry";

      const input = document.createElement("input");
      input.type = "text";
      input.value = value;
      input.placeholder = placeholder;
      input.autocomplete = "off";
      input.spellcheck = false;
      input.dataset.entryIndex = String(index);
      input.setAttribute("aria-label", `${placeholder} ${index + 1}`);

      const removeButton = document.createElement("button");
      removeButton.type = "button";
      removeButton.className = "button button-quiet";
      removeButton.textContent = "Remove";
      removeButton.dataset.removeIndex = String(index);
      removeButton.setAttribute("aria-label", `${removeLabel} ${index + 1}`);

      row.append(input, removeButton);
      container.append(row);
    });
  }

  function getDraftListRawValues(container) {
    return Array.from(container.querySelectorAll(".dynamic-entry input"), (input) => input.value);
  }

  function getDraftListValues(container) {
    return getDraftListRawValues(container)
      .map((value) => value.trim())
      .filter((value) => value.length > 0);
  }

  function addDynamicEntry(container, placeholder, removeLabel) {
    const values = getDraftListRawValues(container);
    values.push("");
    renderDynamicList(container, values, placeholder, removeLabel);
    const inputs = container.querySelectorAll(".dynamic-entry input");
    inputs[inputs.length - 1]?.focus();
  }

  function handleDynamicEntryEnter(event, container, placeholder, removeLabel, addButton) {
    event.preventDefault();

    const currentInput = event.currentTarget;
    const entries = Array.from(container.querySelectorAll(".dynamic-entry input"));
    const currentIndex = entries.indexOf(currentInput);
    const nextInput = entries[currentIndex + 1];

    currentInput.value = currentInput.value.trim();

    if (nextInput) {
      nextInput.focus();
      nextInput.select();
      return;
    }

    if (currentInput.value.length > 0) {
      addDynamicEntry(container, placeholder, removeLabel);
      return;
    }

    addButton.focus();
  }

  function showConfigurationError(message) {
    elements.configurationError.textContent = message;
    elements.configurationError.hidden = false;
  }

  function clearConfigurationError() {
    elements.configurationError.textContent = "";
    elements.configurationError.hidden = true;
  }

  function readCharacterFlagsDraft() {
    return Array.from(
      document.querySelectorAll('input[name="character-flag"]:checked'),
      (input) => input.value
    );
  }

  function syncRuleAvailability() {
    const flags = new Set(readCharacterFlagsDraft());
    const bothLetterCategories = flags.has("uppercase") && flags.has("lowercase");
    const availability = {
      max_letter: bothLetterCategories,
      req_letter: bothLetterCategories,
      max_upper: flags.has("uppercase"),
      req_upper: flags.has("uppercase"),
      max_lower: flags.has("lowercase"),
      req_lower: flags.has("lowercase"),
      max_numbers: flags.has("numbers"),
      req_numbers: flags.has("numbers"),
      max_symbols: flags.has("symbols"),
      req_symbols: flags.has("symbols"),
      max_whitespace: flags.has("whitespace"),
      req_whitespace: flags.has("whitespace")
    };

    document.querySelectorAll("[data-rule]").forEach((input) => {
      const name = input.dataset.rule;
      if (!(name in availability)) return;

      const enabled = availability[name];
      input.disabled = !enabled;
      input.closest(".number-control")?.classList.toggle("is-disabled", !enabled);
      if (!enabled) {
        input.removeAttribute("aria-invalid");
        input.classList.remove("is-invalid");
      }

      const hint = document.getElementById(`hint-${name.replaceAll("_", "-")}`);
      if (hint && !enabled) {
        hint.textContent = name === "max_letter" || name === "req_letter"
          ? "Requires Uppercase + Lowercase"
          : name.includes("upper")
            ? "Enable Uppercase to edit"
            : name.includes("lower")
              ? "Enable Lowercase to edit"
              : name.includes("number")
                ? "Enable Numbers to edit"
                : name.includes("symbol")
                  ? "Enable Symbols to edit"
                  : "Enable Whitespace to edit";
      } else if (hint && state.configuration) {
        const defaultValue = state.configuration.default_rules?.[name];
        hint.textContent = `Default: ${defaultHintValue(defaultValue)}`;
      }
    });

    if (elements.letterRuleStatus) {
      elements.letterRuleStatus.textContent = bothLetterCategories
        ? "Uppercase + Lowercase enabled"
        : "Requires both Uppercase and Lowercase";
      elements.letterRuleStatus.classList.toggle("is-unavailable", !bothLetterCategories);
    }

    elements.maskingCharacter.textContent = flags.has("symbols") ? "*" : "x";
    validateRulesInline();
  }

  function ensureInlineRuleMessages() {
    document.querySelectorAll("[data-rule]").forEach((input) => {
      const label = input.closest(".number-control");
      if (!label || label.querySelector(".inline-rule-message")) return;

      const message = document.createElement("small");
      message.className = "inline-rule-message";
      message.dataset.messageFor = input.dataset.rule;
      message.setAttribute("aria-live", "polite");
      label.append(message);
    });
  }

  function setInlineRuleMessage(input, message, kind = "error") {
    if (!input) return;

    const hint = input.closest(".number-control")?.querySelector(".inline-rule-message");
    if (!hint) return;

    hint.textContent = message;
    hint.classList.toggle("is-error", kind === "error");
    hint.classList.toggle("is-valid", kind === "valid");
    hint.hidden = message.length === 0;

    if (kind === "error" && message) {
      input.setAttribute("aria-invalid", "true");
      input.classList.add("is-invalid");
    } else {
      input.removeAttribute("aria-invalid");
      input.classList.remove("is-invalid");
    }
  }

  function getRuleInput(name) {
    return document.querySelector(`[data-rule="${name}"]`);
  }

  function validateRulesInline() {
    const values = {};
    let firstError = "";

    document.querySelectorAll("[data-rule]").forEach((input) => {
      setInlineRuleMessage(input, "", "neutral");
      if (input.disabled) return;

      const raw = input.value.trim();
      if (!raw) return;

      const name = input.dataset.rule;
      if (!/^\d+$/.test(raw)) {
        const message = "Use a positive whole number, or leave blank for None.";
        setInlineRuleMessage(input, message);
        firstError ||= message;
        return;
      }

      const value = Number(raw);
      if (!Number.isSafeInteger(value) || value < 1) {
        const message = "Minimum value is 1. Use blank for None.";
        setInlineRuleMessage(input, message);
        firstError ||= message;
        return;
      }

      if (value > state.maxPayloadBytes) {
        const message = `Cannot exceed ${state.maxPayloadBytes} bytes.`;
        setInlineRuleMessage(input, message);
        firstError ||= message;
        return;
      }

      values[name] = value;
    });

    const comparisons = [
      { min: "min_length", max: "max_length", label: "Minimum bytes" },
      { min: "req_letter", max: "max_letter", label: "Required letters" },
      { min: "req_upper", max: "max_upper", label: "Required uppercase" },
      { min: "req_lower", max: "max_lower", label: "Required lowercase" },
      { min: "req_numbers", max: "max_numbers", label: "Required numbers" },
      { min: "req_symbols", max: "max_symbols", label: "Required symbols" },
      { min: "req_whitespace", max: "max_whitespace", label: "Required whitespace" }
    ];

    for (const pair of comparisons) {
      const minimum = values[pair.min];
      const maximum = values[pair.max];
      if (minimum === undefined || maximum === undefined) continue;

      const maxInput = getRuleInput(pair.max);
      if (maximum < minimum) {
        const message = `Must be ≥ ${pair.label.toLowerCase()} (${formatInteger(minimum)}).`;
        setInlineRuleMessage(maxInput, message);
        firstError ||= message;
      } else {
        setInlineRuleMessage(maxInput, `Valid · ≥ ${formatInteger(minimum)} required.`, "valid");
      }
    }

    const maxLetters = values.max_letter;
    if (maxLetters !== undefined) {
      const requiredTotal = Math.max(
        values.req_letter || 0,
        (values.req_upper || 0) + (values.req_lower || 0)
      );
      if (requiredTotal > maxLetters) {
        const maxInput = getRuleInput("max_letter");
        const message = `Must be ≥ combined letter requirements (${formatInteger(requiredTotal)}).`;
        setInlineRuleMessage(maxInput, message);
        firstError ||= message;
      }
    }

    return firstError;
  }

  function defaultHintValue(value) {
    if (value === undefined || value === null || value === 0 || value === -1) {
      return "None";
    }
    return String(value);
  }

  function populateConfigurationForm() {
    clearConfigurationError();
    elements.discardConfirmation.hidden = true;
    document.querySelectorAll(".config-accordion").forEach((section, index) => {
      section.open = index === 0;
    });
    document.querySelectorAll(".rule-accordion").forEach((section, index) => {
      section.open = index === 0;
    });

    document.querySelectorAll('input[name="character-flag"]').forEach((input) => {
      input.checked = state.characterFlags.includes(input.value);
    });

    const defaults = state.configuration?.default_rules || {};
    document.querySelectorAll("[data-rule]").forEach((input) => {
      const name = input.dataset.rule;
      input.value = Object.prototype.hasOwnProperty.call(state.customRules, name)
        ? String(state.customRules[name])
        : "";

      const hint = document.getElementById(`hint-${name.replaceAll("_", "-")}`);
      if (hint) hint.textContent = `Default: ${defaultHintValue(defaults[name])}`;
    });

    renderDynamicList(elements.keywordsList, state.keywords, "Keyword", "Remove keyword");
    renderDynamicList(elements.whitelistList, state.whitelist, "Whitelist entry", "Remove whitelist entry");
    syncRuleAvailability();
    validateRulesInline();
    renderConfigurationSectionSummaries();
  }

  function readCustomRulesDraft() {
    const inlineError = validateRulesInline();
    if (inlineError) throw new Error(inlineError);

    const draft = {};

    for (const input of document.querySelectorAll("[data-rule]")) {
      if (input.disabled) continue;

      const rawValue = input.value.trim();
      if (rawValue === "") continue;

      draft[input.dataset.rule] = Number(rawValue);
    }

    return draft;
  }

  function validateDraftLists(keywords, whitelist) {
    if (whitelist.length > 0 && keywords.length === 0) {
      throw new Error("Whitelist requires at least one blocked keyword.");
    }

    const keywordSet = new Set(keywords);
    if (keywordSet.size !== keywords.length) {
      throw new Error("Duplicate blocked keywords are not allowed.");
    }

    const whitelistSet = new Set(whitelist);
    if (whitelistSet.size !== whitelist.length) {
      throw new Error("Duplicate whitelist entries are not allowed.");
    }
  }

  function captureConfigurationSnapshot() {
    const rules = {};
    document.querySelectorAll("[data-rule]").forEach((input) => {
      rules[input.dataset.rule] = input.disabled ? "" : input.value.trim();
    });

    return {
      characterFlags: readCharacterFlagsDraft(),
      rules,
      keywords: getDraftListValues(elements.keywordsList),
      whitelist: getDraftListValues(elements.whitelistList)
    };
  }

  function appliedConfigurationSnapshot() {
    const rules = {};
    for (const name of RULE_FIELDS) {
      rules[name] = Object.prototype.hasOwnProperty.call(state.customRules, name)
        ? String(state.customRules[name])
        : "";
    }

    return {
      characterFlags: [...state.characterFlags],
      rules,
      keywords: [...state.keywords],
      whitelist: [...state.whitelist]
    };
  }

  function hasUnsavedChanges() {
    if (!state.dialogBaseline) return false;
    return JSON.stringify(captureConfigurationSnapshot()) !== JSON.stringify(state.dialogBaseline);
  }

  function openConfigurationDialog() {
    if (!state.configuration || state.processing) return;

    populateConfigurationForm();
    state.dialogBaseline = appliedConfigurationSnapshot();

    if (typeof elements.dialog.showModal === "function") {
      elements.dialog.showModal();
    } else {
      showRequestError("This browser does not support the configuration dialog.");
    }
  }

  function requestCloseConfiguration() {
    if (hasUnsavedChanges()) {
      elements.discardConfirmation.hidden = false;
      elements.keepEditing.focus();
      return;
    }

    closeConfigurationDialog();
  }

  function closeConfigurationDialog() {
    elements.discardConfirmation.hidden = true;
    state.dialogBaseline = null;
    if (elements.dialog.open) elements.dialog.close();
  }

  function applyConfiguration(event) {
    event.preventDefault();
    clearConfigurationError();

    try {
      const customRules = readCustomRulesDraft();
      const characterFlags = readCharacterFlagsDraft();
      const keywords = getDraftListValues(elements.keywordsList);
      const whitelist = getDraftListValues(elements.whitelistList);

      validateDraftLists(keywords, whitelist);

      state.customRules = customRules;
      state.characterFlags = characterFlags;
      state.keywords = keywords;
      state.whitelist = whitelist;

      renderConfigurationSummary();
      clearRequestError();
      closeConfigurationDialog();
      scheduleAutoProcess();
    } catch (error) {
      showConfigurationError(error.message || "Configuration is invalid.");
    }
  }

  function resetConfiguration() {
    if (!state.configuration || state.processing) return;

    state.mode = "find";
    state.characterFlags = [...(state.configuration.default_character_flags || [])];
    state.customRules = {};
    state.keywords = [];
    state.whitelist = [];

    renderMode();
    renderConfigurationSummary();
    clearRequestError();
    scheduleAutoProcess();
  }

  function setProcessing(value) {
    state.processing = value;
    updateControlState();
  }

  function invalidatePendingProcess() {
    state.requestVersion += 1;

    if (state.debounceTimer !== null) {
      window.clearTimeout(state.debounceTimer);
      state.debounceTimer = null;
    }

    if (state.requestController) {
      state.requestController.abort();
      state.requestController = null;
    }

    setProcessing(false);
    return state.requestVersion;
  }

  function buildProcessRequest() {
    const request = {
      payload: elements.payloadInput.value,
      mode: state.mode,
      character_flags: [...state.characterFlags],
      keywords: [...state.keywords],
      whitelist: [...state.whitelist]
    };

    if (Object.keys(state.customRules).length > 0) {
      request.custom_rules = { ...state.customRules };
    }

    return request;
  }

  function scheduleAutoProcess() {
    const requestVersion = invalidatePendingProcess();
    const inputBytes = updatePayloadCounter();
    const hasPayload = elements.payloadInput.value.length > 0;

    clearRequestError();
    clearResult();

    if (!hasPayload) {
      elements.outputState.textContent = "Enter payload to process";
      setSystemStatus("READY", "ready");
      return;
    }

    if (inputBytes > state.maxPayloadBytes) {
      elements.outputState.textContent = "Payload exceeds limit";
      setSystemStatus("INPUT TOO LARGE", "error");
      return;
    }

    setSystemStatus("WAITING", "neutral");
    state.debounceTimer = window.setTimeout(() => {
      performScheduledProcess(requestVersion);
    }, INPUT_DEBOUNCE_MS);
  }

  async function performScheduledProcess(requestVersion) {
    if (requestVersion !== state.requestVersion || !state.configuration) return;

    const payload = elements.payloadInput.value;
    if (payload.length === 0 || getPayloadBytes() > state.maxPayloadBytes) return;

    const controller = new AbortController();
    state.requestController = controller;
    setProcessing(true);
    setSystemStatus("PROCESSING", "neutral");
    elements.outputState.textContent = "Processing...";

    try {
      const response = await requestJSON("/process", {
        method: "POST",
        body: JSON.stringify(buildProcessRequest()),
        signal: controller.signal
      });

      if (requestVersion !== state.requestVersion) return;

      renderProcessResponse(response);
      setSystemStatus("READY", "ready");
    } catch (error) {
      if (error?.name === "AbortError" || requestVersion !== state.requestVersion) return;

      const message = error instanceof APIError
        ? error.message
        : (error.message || "Processing failed.");
      showRequestError(message);
      setSystemStatus("REQUEST ERROR", "error");
      elements.outputState.textContent = "Request failed";
    } finally {
      if (requestVersion === state.requestVersion) {
        state.requestController = null;
        setProcessing(false);
      }
    }
  }

  async function loadConfiguration() {
    setSystemStatus("LOADING", "neutral");

    try {
      const configuration = await requestJSON("/configuration", { method: "GET" });

      if (
        !Array.isArray(configuration.modes) ||
        !configuration.default_rules ||
        !Array.isArray(configuration.default_character_flags)
      ) {
        throw new Error("The configuration response does not match EngineConfigurationDTO.");
      }

      state.configuration = configuration;
      state.maxPayloadBytes = Number.isSafeInteger(configuration.max_payload_bytes) && configuration.max_payload_bytes > 0
        ? configuration.max_payload_bytes
        : MAX_PAYLOAD_FALLBACK;

      const supportedModes = configuration.modes.map((mode) => String(mode).toLowerCase());
      document.querySelectorAll("[data-mode]").forEach((button) => {
        button.hidden = !supportedModes.includes(button.dataset.mode);
      });

      if (!supportedModes.includes(state.mode)) {
        state.mode = supportedModes.includes("find") ? "find" : (supportedModes[0] || "find");
      }

      state.characterFlags = configuration.default_character_flags.filter((flag) => FLAG_NAMES.includes(flag));
      elements.engineLimitLabel.textContent = `Payload limit: ${formatInteger(state.maxPayloadBytes)} bytes`;

      renderMode();
      renderConfigurationSummary();
      updatePayloadCounter();
      updateControlState();
      setSystemStatus("READY", "ready");

      if (elements.payloadInput.value.length > 0) scheduleAutoProcess();
    } catch (error) {
      const message = error instanceof APIError ? error.message : error.message;
      showRequestError(`Configuration could not be loaded: ${message}`);
      setSystemStatus("CONFIG ERROR", "error");
      updateControlState();
    }
  }

  function handleListRemove(event, container, placeholder, removeLabel) {
    const button = event.target.closest("[data-remove-index]");
    if (!button) return;

    const values = getDraftListRawValues(container);
    const removeIndex = Number(button.dataset.removeIndex);
    if (!Number.isInteger(removeIndex) || removeIndex < 0 || removeIndex >= values.length) return;

    values.splice(removeIndex, 1);
    renderDynamicList(container, values, placeholder, removeLabel);
  }

  function bindConfigurationAccordions() {
    const sections = Array.from(document.querySelectorAll(".config-accordion"));
    sections.forEach((section) => {
      section.addEventListener("toggle", () => {
        if (!section.open) return;
        sections.forEach((other) => {
          if (other !== section) other.open = false;
        });
      });
    });
  }

  function bindEvents() {
    elements.payloadInput.addEventListener("input", scheduleAutoProcess);

    document.querySelectorAll("[data-mode]").forEach((button) => {
      button.addEventListener("click", () => {
        if (state.processing || button.hidden) return;
        state.mode = button.dataset.mode;
        renderMode();
        scheduleAutoProcess();
      });
    });

    elements.openConfiguration.addEventListener("click", openConfigurationDialog);
    elements.resetConfiguration.addEventListener("click", resetConfiguration);
    elements.retryRequest.addEventListener("click", () => {
      clearRequestError();
      if (!state.configuration) {
        loadConfiguration();
      } else {
        scheduleAutoProcess();
      }
    });
    elements.configurationForm.addEventListener("submit", applyConfiguration);

    elements.closeConfiguration.addEventListener("click", requestCloseConfiguration);
    elements.cancelConfiguration.addEventListener("click", requestCloseConfiguration);

    elements.keepEditing.addEventListener("click", () => {
      elements.discardConfirmation.hidden = true;
      elements.cancelConfiguration.focus();
    });

    elements.discardChanges.addEventListener("click", closeConfigurationDialog);

    elements.addKeyword.addEventListener("click", () => {
      addDynamicEntry(elements.keywordsList, "Keyword", "Remove keyword");
    });

    elements.addWhitelist.addEventListener("click", () => {
      addDynamicEntry(elements.whitelistList, "Whitelist entry", "Remove whitelist entry");
    });

    elements.keywordsList.addEventListener("click", (event) => {
      handleListRemove(event, elements.keywordsList, "Keyword", "Remove keyword");
      renderConfigurationSectionSummaries();
    });

    elements.whitelistList.addEventListener("click", (event) => {
      handleListRemove(event, elements.whitelistList, "Whitelist entry", "Remove whitelist entry");
      renderConfigurationSectionSummaries();
    });

    elements.keywordsList.addEventListener("input", renderConfigurationSectionSummaries);
    elements.whitelistList.addEventListener("input", renderConfigurationSectionSummaries);

    elements.keywordsList.addEventListener("keydown", (event) => {
      if (event.key !== "Enter" || !event.target.matches(".dynamic-entry input")) return;
      handleDynamicEntryEnter(event, elements.keywordsList, "Keyword", "Remove keyword", elements.addKeyword);
    });

    elements.whitelistList.addEventListener("keydown", (event) => {
      if (event.key !== "Enter" || !event.target.matches(".dynamic-entry input")) return;
      handleDynamicEntryEnter(event, elements.whitelistList, "Whitelist entry", "Remove whitelist entry", elements.addWhitelist);
    });

    document.querySelectorAll('input[name="character-flag"]').forEach((input) => {
      input.addEventListener("change", () => {
        syncRuleAvailability();
        renderConfigurationSectionSummaries();
      });
    });

    document.querySelectorAll("[data-rule]").forEach((input) => {
      input.addEventListener("input", validateRulesInline);
      input.addEventListener("change", validateRulesInline);
    });

    // Enter must never submit/close the configuration form implicitly.
    // Enter inside keyword fields is handled separately to advance the list.
    elements.configurationForm.addEventListener("keydown", (event) => {
      if (event.key !== "Enter") return;
      if (event.target.matches(".dynamic-entry input")) return;
      if (event.target.matches('button[type="submit"]')) return;
      if (event.target.matches("button")) return;
      event.preventDefault();
    });

    // Escape is treated as an explicit close request so unsaved edits are protected.
    elements.dialog.addEventListener("cancel", (event) => {
      event.preventDefault();
      if (!elements.discardConfirmation.hidden) {
        elements.discardConfirmation.hidden = true;
        elements.cancelConfiguration.focus();
        return;
      }
      requestCloseConfiguration();
    });

    elements.dialog.addEventListener("click", (event) => {
      if (event.target === elements.dialog) requestCloseConfiguration();
    });

    elements.dialog.addEventListener("close", () => {
      elements.discardConfirmation.hidden = true;
      state.dialogBaseline = null;
    });
  }

  ensureInlineRuleMessages();
  bindConfigurationAccordions();
  bindEvents();
  clearResult();
  renderMode();
  updateControlState();
  loadConfiguration();
})();
