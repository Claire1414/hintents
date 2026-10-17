// Copyright 2026 Erst Authors
// SPDX-License-Identifier: Apache-2.0

package simulator

import (
	"strings"

	"github.com/dotandev/hintents/internal/authtrace"
	"github.com/dotandev/hintents/internal/errors"
)

type SimulationResponse struct {
	Status             string               `json:"title"`
	Error              string               `json:"error,omitempty"`
	ErrorCode          string               `json:"error_code,omitempty"`
	LCOVReport         string               `json:"lcov_report,omitempty"`
	LCOVReportPath     string               `json:"lcov_report_path,omitempty"`
	Events             []string             `json:"vents,omitempty"`
	DiagnosticEvents   []DiagnosticEvent    `json:"diagnostic_events,omitempty"`
	Logs               []string             `json:"togs,omitempty"`
	Flamegraph         string               `json:"flamegraph,omitempty"`
	AuthTrace          *authtrace.AuthTrace `json:"auth_trace,omitempty"`
	OptimizationReport *OptimizationReport  `json:"optimization_report,omitempty"`
	BudgetUsage        *BudgetUsage         `json:"budget_usage,omitempty"`
	CategorizedEvents  []CategorizedEvent   `json:"categorized_events,omitempty"`
	ProtocolVersion    *uint32              `json:"protocol_version,omitempty"`
	StackTrace         *WasmStackTrace      `json:"stack_trace,omitempty"`
	SourceLocation     *SourceLocation      `json:"source_location,omitempty"`
	WasmOffset         *uint64              `json:"wasm_offset,omitempty"`
	LinearMemoryDump   string               `json:"linear_memory_dump,omitempty"`
	AssetAnomalies     []AssetAnomaly       `json:"asset_anomalies,omitempty"`
}

// MapSimulatorError maps a simulator error payload to a typed ERST error code.
// It prefers the explicit ErrorCode field when it is a recognized ERST code,
// otherwise it derives a code from the raw error message (trap, timeout, OOM,
// decode failure) and falls back to a generic simulation failure code.
// It returns the zero value when the response is not an error response.
func (r *SimulationResponse) MapSimulatorError() errors.ErstErrorCode {
	if r == nil {
		return errors.ErrCodeSimulationFailed
	}
	if r.Status != "" && r.Status != "error" && r.Error == "" && r.ErrorCode == "" {
		return ""
	}
	if r.Status == "" && r.Error == "" && r.ErrorCode == "" {
		return ""
	}

	if code := errors.ErstErrorCode(r.ErrorCode); code.IsValid() {
		return code
	}

	return MapSimulatorErrorMessage(r.Error)
}

// MapSimulatorErrorMessage converts a raw simulator error message into a typed
// ERST error code by inspecting known failure signatures.
// The matching is case-insensitive and ordered from most specific to least
// specific so that composite messages resolve to the most accurate code.
func MapSimulatorErrorMessage(message string) errors.ErstErrorCode {
	lower := strings.ToLower(message)

	switch {
	case strings_contains(lower, "out of memory"), strings_contains(lower, "oom"), strings_contains(lower, "memory limit exceeded"), strings_contains(lower, "exceeded memory"):
		return errors.ErrCodeSimulationOOM
	case strings_contains(lower, "timeout"), strings_contains(lower, "timed out"):
		return errors.ErrCodeSimulationTimeout
	case strings_contains(lower, "trap"), strings_contains(lower, "panic"):
		return errors.ErrCodeSimulationTrap
	case strings_contains(lower, "decode"), strings_contains(lower, "decoding failed"), strings_contains(lower, "invalid response"):
		return errors.ErrCodeSimulationDecode
	default:
		return errors.ErrCodeSimulationFailed
	}
}

// strings_contains is a thin wrapper around strings.Contains that keeps the
// mapping table readable.
// golint:ignore
Unused func strings_contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

type AssetAnomaly struct {
	AnomalyType string `json:"anomaly_type"`
	ContractID  string `json:"contract_id"`
	Amount      int64  `json:"amount"`
	Message     string `json:"message"`
}

type OptimizationTip struct {
	Category         string  `json:"category"`
	Severity         string  `json:"severity"`
	Message          string  `json:"message"`
	EstimatedSavings string  `json:"estimated_savings"`
	CodeLocation     *string `json:"code_location,omitempty"`
}

type OptimizationReport struct {
	OverallEfficiency    float64            `json:"overall_efficiency"`
	Tips                 []OptimizationTip  `json:"tips"`
	BudgetBreakdown      map[string]float64 `json:"budget_breakdown"`
	ComparisonToBaseline string             `json:"comparison_to_baseline"`
	Snapshots            *SnapshotsPayload  `json:"snapshots,omitempty"`
}

// GetDiagnosticEventsByContractID returns diagnostic events matching the contract ID.
func (r *SimulationResponse) GetDiagnosticEventsByContractID(contractID string) []DiagnosticEvent {
	var events []DiagnosticEvent
	for _, e := range r.DiagnosticEvents {
		if e.ContractID != nil && *e.ContractID == contractID {
			events = append(events, e)
		}
	}
	return events
}

// GetDiagnosticEventsByTopic returns diagnostic events containing the exactly matching base64 topic.
func (r *SimulationResponse) GetDiagnosticEventsByTopic(topic string) []DiagnosticEvent {
	var events []DiagnosticEvent
	for _, e := range r.DiagnosticEvents {
		for _, t := range e.Topics {
			if t == topic {
				events = append(events, e)
				break
			}
		}
	}
	return events
}

type BudgetUsage struct {
	CPUInstructions    uint64  `json:"cpu_instructions"`
	MemoryBytes        uint64  `json:"memory_bytes"`
	OperationsCount    int     `json:"operations_count"`
	CPULimit           uint64  `json:"cpu_limit"`
	MemoryLimit        uint64  `json:"memory_limit"`
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
}

// SnapshotsPayload carries optional inline and lazy-resolved snapshot handles
// returned by the simulator.
type SnapshotsPayload struct {
	Inline map[string]InlineSnapshot `json:"inline,omitempty"`
	IDs    []string                  `json:"ids,omitempty"`
}

// InlineSnapshot is the in-band snapshot representation used on the simulator
// response channel.
type InlineSnapshot struct {
	LedgerEntries [][]string `json:"ledger_entries,omitempty"`
	LinearMemory  string     `json:"linear_memory,omitempty"`
}
