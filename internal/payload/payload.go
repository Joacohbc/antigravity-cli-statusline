package payload

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

const (
	// MaxPayloadSize defines the upper limit of payload ingestion (2MB).
	MaxPayloadSize = 2 * 1024 * 1024

	// DefaultTerminalWidth is the fallback terminal width when unspecified or invalid.
	DefaultTerminalWidth = 80

	// DefaultAgentState is the fallback agent state when unspecified.
	DefaultAgentState = "idle"
)

// ModelInfo describes the active LLM model.
type ModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// GetID returns the sanitized model ID, or "unknown" if empty.
func (m ModelInfo) GetID() string {
	return cmp.Or(Sanitize(m.ID), "unknown")
}

// GetDisplayName returns the sanitized display name, or falls back to GetID().
func (m ModelInfo) GetDisplayName() string {
	return cmp.Or(Sanitize(m.DisplayName), m.GetID())
}

// Name returns the best display identifier for the model.
func (m ModelInfo) Name() string {
	return m.GetDisplayName()
}

// WorkspaceInfo describes the active directory and project paths.
type WorkspaceInfo struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

// GetCurrentDir returns the sanitized current working directory.
func (w WorkspaceInfo) GetCurrentDir() string {
	return Sanitize(w.CurrentDir)
}

// GetProjectDir returns the sanitized project directory.
func (w WorkspaceInfo) GetProjectDir() string {
	return Sanitize(w.ProjectDir)
}

// TokenUsage contains granular token accounting details.
type TokenUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// GetInputTokens returns the input token count (guaranteed >= 0).
func (t TokenUsage) GetInputTokens() int {
	return max(0, t.InputTokens)
}

// GetOutputTokens returns the output token count (guaranteed >= 0).
func (t TokenUsage) GetOutputTokens() int {
	return max(0, t.OutputTokens)
}

// GetCacheCreationInputTokens returns the cache creation token count (guaranteed >= 0).
func (t TokenUsage) GetCacheCreationInputTokens() int {
	return max(0, t.CacheCreationInputTokens)
}

// GetCacheReadInputTokens returns the cache read token count (guaranteed >= 0).
func (t TokenUsage) GetCacheReadInputTokens() int {
	return max(0, t.CacheReadInputTokens)
}

// TotalTokens returns the sum of input and output tokens.
func (t TokenUsage) TotalTokens() int {
	return t.GetInputTokens() + t.GetOutputTokens()
}

// ContextWindowInfo contains context usage and token limits.
type ContextWindowInfo struct {
	TotalInputTokens    int        `json:"total_input_tokens"`
	TotalOutputTokens   int        `json:"total_output_tokens"`
	ContextWindowSize   int        `json:"context_window_size"`
	UsedPercentage      float64    `json:"used_percentage"`
	RemainingPercentage float64    `json:"remaining_percentage"`
	CurrentUsage        TokenUsage `json:"current_usage"`
}

// ContextInfo is an alias for ContextWindowInfo for cross-package compatibility.
type ContextInfo = ContextWindowInfo

// GetTotalInputTokens returns the total input tokens (guaranteed >= 0).
func (c ContextWindowInfo) GetTotalInputTokens() int {
	return max(0, c.TotalInputTokens)
}

// GetTotalOutputTokens returns the total output tokens (guaranteed >= 0).
func (c ContextWindowInfo) GetTotalOutputTokens() int {
	return max(0, c.TotalOutputTokens)
}

// GetContextWindowSize returns the configured context window capacity.
func (c ContextWindowInfo) GetContextWindowSize() int {
	return max(0, c.ContextWindowSize)
}

// GetUsedPercentage returns the context window used percentage clamped to [0.0, 100.0].
func (c ContextWindowInfo) GetUsedPercentage() float64 {
	if c.UsedPercentage < 0.0 {
		return 0.0
	}
	if c.UsedPercentage > 100.0 {
		return 100.0
	}
	return c.UsedPercentage
}

// GetRemainingPercentage returns the remaining percentage clamped to [0.0, 100.0].
func (c ContextWindowInfo) GetRemainingPercentage() float64 {
	if c.RemainingPercentage == 0.0 && c.UsedPercentage <= 0.0 {
		return 100.0
	}
	if c.RemainingPercentage == 0.0 && c.UsedPercentage > 0.0 {
		return max(0.0, 100.0-c.UsedPercentage)
	}
	if c.RemainingPercentage < 0.0 {
		return 0.0
	}
	if c.RemainingPercentage > 100.0 {
		return 100.0
	}
	return c.RemainingPercentage
}

// GetCurrentUsage returns the current TokenUsage struct safely.
func (c ContextWindowInfo) GetCurrentUsage() TokenUsage {
	return c.CurrentUsage
}

// VCSInfo contains version control state.
type VCSInfo struct {
	Type   string `json:"type"`
	Branch string `json:"branch"`
	Client string `json:"client"`
	Dirty  bool   `json:"dirty"`
}

// GetType returns the sanitized VCS type, defaulting to "git".
func (v VCSInfo) GetType() string {
	return cmp.Or(Sanitize(v.Type), "git")
}

// GetBranch returns the sanitized VCS branch name.
func (v VCSInfo) GetBranch() string {
	return Sanitize(v.Branch)
}

// GetClient returns the sanitized VCS client identifier.
func (v VCSInfo) GetClient() string {
	return Sanitize(v.Client)
}

// IsDirty reports whether there are uncommitted working tree changes.
func (v VCSInfo) IsDirty() bool {
	return v.Dirty
}

// SandboxInfo describes environment execution isolation.
type SandboxInfo struct {
	Enabled      bool `json:"enabled"`
	AllowNetwork bool `json:"allow_network"`
}

// IsEnabled reports whether sandboxing is active.
func (s SandboxInfo) IsEnabled() bool {
	return s.Enabled
}

// IsAllowNetwork reports whether the sandbox permits outbound network access.
func (s SandboxInfo) IsAllowNetwork() bool {
	return s.AllowNetwork
}

// QuotaDetail contains rate limit metrics and reset countdowns.
type QuotaDetail struct {
	RemainingFraction float64 `json:"remaining_fraction"`
	UsedPercentage    float64 `json:"used_percentage,omitempty"`
	ResetTime         string  `json:"reset_time"`
	ResetInSeconds    int     `json:"reset_in_seconds"`
}

// RateLimit describes duration-based limits (e.g. five_hour, seven_day).
type RateLimit struct {
	UsedPercentage      float64 `json:"used_percentage"`
	RemainingPercentage float64 `json:"remaining_percentage"`
	ResetTime           string  `json:"reset_time"`
	ResetInSeconds      int     `json:"reset_in_seconds"`
}

// GetRemainingFraction returns the remaining quota fraction.
func (q QuotaDetail) GetRemainingFraction() float64 {
	if q.RemainingFraction == 0.0 && q.ResetTime == "" && q.ResetInSeconds == 0 {
		return 1.0
	}
	if q.RemainingFraction < 0.0 {
		return 0.0
	}
	if q.RemainingFraction > 1.0 {
		return 1.0
	}
	return q.RemainingFraction
}

// GetResetTime returns the sanitized reset timestamp string.
func (q QuotaDetail) GetResetTime() string {
	return Sanitize(q.ResetTime)
}

// GetResetInSeconds returns the seconds until reset (guaranteed >= 0).
func (q QuotaDetail) GetResetInSeconds() int {
	return max(0, q.ResetInSeconds)
}

// QuotaInfo is a type alias for the Quota map.
type QuotaInfo = map[string]QuotaDetail

// VimInfo contains Vim/Neovim statusline integration state.
type VimInfo struct {
	Mode string `json:"mode"`
}

// GetMode returns the sanitized Vim mode.
func (v *VimInfo) GetMode() string {
	if v == nil {
		return ""
	}
	return Sanitize(v.Mode)
}

// StatsInfo summarizes execution metrics for easy component consumption.
type StatsInfo struct {
	ArtifactCount int `json:"artifact_count"`
	TaskCount     int `json:"task_count"`
	SubagentCount int `json:"subagent_count"`
}

// GetArtifactCount returns the number of generated artifacts.
func (s StatsInfo) GetArtifactCount() int {
	return max(0, s.ArtifactCount)
}

// GetTaskCount returns the number of active tasks.
func (s StatsInfo) GetTaskCount() int {
	return max(0, s.TaskCount)
}

// GetSubagentCount returns the number of running subagents.
func (s StatsInfo) GetSubagentCount() int {
	return max(0, s.SubagentCount)
}

// StatePayload is the root JSON structure passed into antigravity-cli-statusline.
type StatePayload struct {
	CWD                     string                 `json:"cwd"`
	SessionID               string                 `json:"session_id"`
	ConversationID          string                 `json:"conversation_id"`
	TranscriptPath          string                 `json:"transcript_path"`
	Model                   ModelInfo              `json:"model"`
	Workspace               WorkspaceInfo          `json:"workspace"`
	Version                 string                 `json:"version"`
	ContextWindow           ContextWindowInfo      `json:"context_window"`
	Exceeds200kTokens       *bool                  `json:"exceeds_200k_tokens"`
	Product                 string                 `json:"product"`
	Quota                   map[string]QuotaDetail `json:"quota"`
	WeeklyUsagePercent      *float64               `json:"weekly_usage_percent,omitempty"`
	SessionUsagePercent     *float64               `json:"session_usage_percent,omitempty"`
	RateLimits              map[string]RateLimit   `json:"rate_limits,omitempty"`
	AgentState              string                 `json:"agent_state"`
	VCS                     VCSInfo                `json:"vcs"`
	Sandbox                 SandboxInfo            `json:"sandbox"`
	ArtifactCount           int                    `json:"artifact_count"`
	PlanTier                string                 `json:"plan_tier"`
	Email                   string                 `json:"email"`
	PendingInputCount       int                    `json:"pending_input_count"`
	ToolConfirmationPending bool                   `json:"tool_confirmation_pending"`
	TaskCount               int                    `json:"task_count"`
	SubagentCount           int                    `json:"-"`
	RawSubagents            json.RawMessage        `json:"subagents"`
	TerminalWidth           int                    `json:"terminal_width"`
	ExecutionMode           string                 `json:"execution_mode"`
	Vim                     *VimInfo               `json:"vim"`
	RawMap                  map[string]json.RawMessage `json:"-"`
}

// NewDefaultStatePayload initializes a safe, default StatePayload instance.
func NewDefaultStatePayload() *StatePayload {
	return &StatePayload{
		AgentState:    DefaultAgentState,
		TerminalWidth: DefaultTerminalWidth,
		Model: ModelInfo{
			ID:          "unknown",
			DisplayName: "Unknown",
		},
		Quota: make(map[string]QuotaDetail),
	}
}

// GetCWD returns the sanitized current working directory.
func (p *StatePayload) GetCWD() string {
	if p == nil {
		return ""
	}
	return Sanitize(p.CWD)
}

// GetSessionID returns the sanitized session identifier.
func (p *StatePayload) GetSessionID() string {
	if p == nil {
		return ""
	}
	return Sanitize(p.SessionID)
}

// GetConversationID returns the sanitized conversation identifier.
func (p *StatePayload) GetConversationID() string {
	if p == nil {
		return ""
	}
	return Sanitize(p.ConversationID)
}

// GetTranscriptPath returns the sanitized transcript file path.
func (p *StatePayload) GetTranscriptPath() string {
	if p == nil {
		return ""
	}
	return Sanitize(p.TranscriptPath)
}

// GetModel returns the ModelInfo struct safely.
func (p *StatePayload) GetModel() ModelInfo {
	if p == nil {
		return ModelInfo{ID: "unknown", DisplayName: "Unknown"}
	}
	return p.Model
}

// GetWorkspace returns the WorkspaceInfo struct safely.
func (p *StatePayload) GetWorkspace() WorkspaceInfo {
	if p == nil {
		return WorkspaceInfo{}
	}
	return p.Workspace
}

// GetVersion returns the sanitized version string.
func (p *StatePayload) GetVersion() string {
	if p == nil {
		return ""
	}
	return Sanitize(p.Version)
}

// GetContextWindow returns the ContextWindowInfo struct safely.
func (p *StatePayload) GetContextWindow() ContextWindowInfo {
	if p == nil {
		return ContextWindowInfo{}
	}
	return p.ContextWindow
}

// GetProduct returns the sanitized product identifier.
func (p *StatePayload) GetProduct() string {
	if p == nil {
		return ""
	}
	return Sanitize(p.Product)
}

// GetQuota returns a defensive clone of the Quota map.
func (p *StatePayload) GetQuota() map[string]QuotaDetail {
	if p == nil || p.Quota == nil {
		return make(map[string]QuotaDetail)
	}
	return maps.Clone(p.Quota)
}

// GetAgentState returns the sanitized agent state, defaulting to "idle".
func (p *StatePayload) GetAgentState() string {
	if p == nil {
		return DefaultAgentState
	}
	return cmp.Or(Sanitize(p.AgentState), DefaultAgentState)
}

// GetVCS returns the VCSInfo struct safely.
func (p *StatePayload) GetVCS() VCSInfo {
	if p == nil {
		return VCSInfo{Type: "git"}
	}
	return p.VCS
}

// GetSandbox returns the SandboxInfo struct safely.
func (p *StatePayload) GetSandbox() SandboxInfo {
	if p == nil {
		return SandboxInfo{}
	}
	return p.Sandbox
}

// GetArtifactCount returns the number of artifacts (guaranteed >= 0).
func (p *StatePayload) GetArtifactCount() int {
	if p == nil {
		return 0
	}
	return max(0, p.ArtifactCount)
}

// GetPlanTier returns the sanitized plan tier name.
func (p *StatePayload) GetPlanTier() string {
	if p == nil {
		return ""
	}
	return Sanitize(p.PlanTier)
}

// GetEmail returns the sanitized email address.
func (p *StatePayload) GetEmail() string {
	if p == nil {
		return ""
	}
	return Sanitize(p.Email)
}

// GetPendingInputCount returns the count of pending input requests (guaranteed >= 0).
func (p *StatePayload) GetPendingInputCount() int {
	if p == nil {
		return 0
	}
	return max(0, p.PendingInputCount)
}

// IsToolConfirmationPending reports whether user tool confirmation is pending.
func (p *StatePayload) IsToolConfirmationPending() bool {
	if p == nil {
		return false
	}
	return p.ToolConfirmationPending
}

// GetTaskCount returns the number of active tasks (guaranteed >= 0).
func (p *StatePayload) GetTaskCount() int {
	if p == nil {
		return 0
	}
	return max(0, p.TaskCount)
}

// GetSubagentCount returns the number of running subagents (guaranteed >= 0).
func (p *StatePayload) GetSubagentCount() int {
	if p == nil {
		return 0
	}
	return max(0, p.SubagentCount)
}

// GetTerminalWidth returns the terminal width, falling back to 80 if invalid.
func (p *StatePayload) GetTerminalWidth() int {
	if p == nil || p.TerminalWidth <= 0 {
		return DefaultTerminalWidth
	}
	return p.TerminalWidth
}

// GetExecutionMode returns the sanitized execution mode.
func (p *StatePayload) GetExecutionMode() string {
	if p == nil {
		return ""
	}
	return Sanitize(p.ExecutionMode)
}

// GetVim returns the VimInfo struct pointer or nil safely.
func (p *StatePayload) GetVim() *VimInfo {
	if p == nil {
		return nil
	}
	return p.Vim
}

// GetStats returns a consolidated StatsInfo struct.
func (p *StatePayload) GetStats() StatsInfo {
	if p == nil {
		return StatsInfo{}
	}
	return StatsInfo{
		ArtifactCount: p.GetArtifactCount(),
		TaskCount:     p.GetTaskCount(),
		SubagentCount: p.GetSubagentCount(),
	}
}

// SanitizeFields sanitizes all string fields within the payload in-place.
func (p *StatePayload) SanitizeFields() {
	if p == nil {
		return
	}
	p.CWD = Sanitize(p.CWD)
	p.SessionID = Sanitize(p.SessionID)
	p.ConversationID = Sanitize(p.ConversationID)
	p.TranscriptPath = Sanitize(p.TranscriptPath)
	p.Model.ID = Sanitize(p.Model.ID)
	p.Model.DisplayName = Sanitize(p.Model.DisplayName)
	p.Workspace.CurrentDir = Sanitize(p.Workspace.CurrentDir)
	p.Workspace.ProjectDir = Sanitize(p.Workspace.ProjectDir)
	p.Version = Sanitize(p.Version)
	p.Product = Sanitize(p.Product)
	p.AgentState = Sanitize(p.AgentState)
	p.VCS.Type = Sanitize(p.VCS.Type)
	p.VCS.Branch = Sanitize(p.VCS.Branch)
	p.VCS.Client = Sanitize(p.VCS.Client)
	p.PlanTier = Sanitize(p.PlanTier)
	p.Email = Sanitize(p.Email)
	p.ExecutionMode = Sanitize(p.ExecutionMode)
	if p.Vim != nil {
		p.Vim.Mode = Sanitize(p.Vim.Mode)
	}
}

// resolveSubagentCount parses the flexible subagents field (either integer or array).
func (p *StatePayload) resolveSubagentCount() {
	if p == nil || len(p.RawSubagents) == 0 {
		return
	}

	var count int
	if err := json.Unmarshal(p.RawSubagents, &count); err == nil {
		p.SubagentCount = max(0, count)
		return
	}

	var list []any
	if err := json.Unmarshal(p.RawSubagents, &list); err == nil {
		p.SubagentCount = len(list)
	}
}

// Parse reads and unmarshals a StatePayload from an io.Reader, enforcing a 2MB maximum limit.
// It returns a safe default StatePayload even when parsing errors occur.
func Parse(reader io.Reader) (*StatePayload, error) {
	if reader == nil {
		return NewDefaultStatePayload(), &ParseError{
			Op:  "read",
			Err: ErrEmptyPayload,
			Msg: "reader is nil",
		}
	}

	// Bounded ingestion: Read up to MaxPayloadSize + 1 bytes to detect truncation / oversized payload
	limitedReader := io.LimitReader(reader, int64(MaxPayloadSize)+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return NewDefaultStatePayload(), &ParseError{
			Op:  "read",
			Err: err,
			Msg: "failed reading from stream",
		}
	}

	if len(data) > MaxPayloadSize {
		return NewDefaultStatePayload(), &ParseError{
			Op:  "limit_check",
			Err: ErrPayloadOversized,
			Msg: fmt.Sprintf("payload exceeds %d bytes limit", MaxPayloadSize),
		}
	}

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return NewDefaultStatePayload(), &ParseError{
			Op:  "read",
			Err: ErrEmptyPayload,
			Msg: "payload is empty or whitespace",
		}
	}

	return ParseBytes(trimmed)
}

// ParseBytes parses a raw byte slice into a StatePayload.
func ParseBytes(data []byte) (*StatePayload, error) {
	if len(data) > MaxPayloadSize {
		return NewDefaultStatePayload(), &ParseError{
			Op:  "limit_check",
			Err: ErrPayloadOversized,
			Msg: fmt.Sprintf("payload exceeds %d bytes limit", MaxPayloadSize),
		}
	}

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return NewDefaultStatePayload(), &ParseError{
			Op:  "parse",
			Err: ErrEmptyPayload,
			Msg: "payload is empty or whitespace",
		}
	}

	payload := &StatePayload{
		AgentState:    DefaultAgentState,
		TerminalWidth: DefaultTerminalWidth,
		Quota:         make(map[string]QuotaDetail),
	}

	if err := json.Unmarshal(trimmed, payload); err != nil {
		return payload, &ParseError{
			Op:  "unmarshal",
			Err: ErrInvalidJSON,
			Msg: err.Error(),
		}
	}

	payload.SanitizeFields()
	payload.resolveSubagentCount()
	payload.resolveVCS()

	// Apply fallback defaults if unmarshaling left them blank
	payload.AgentState = cmp.Or(payload.AgentState, DefaultAgentState)
	if payload.TerminalWidth <= 0 {
		payload.TerminalWidth = DefaultTerminalWidth
	}
	if payload.Quota == nil {
		payload.Quota = make(map[string]QuotaDetail)
	}

	var rawMap map[string]json.RawMessage
	if json.Unmarshal(trimmed, &rawMap) == nil {
		payload.RawMap = rawMap
	}

	return payload, nil
}

// resolveVCS detects Git branch locally if omitted in the payload.
func (p *StatePayload) resolveVCS() {
	if p.VCS.Branch != "" {
		return
	}
	// Check directories in priority order: CWD, Workspace.CurrentDir, Workspace.ProjectDir
	dirs := []string{p.CWD, p.Workspace.CurrentDir, p.Workspace.ProjectDir}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		if branch, dirty := findGitBranch(dir); branch != "" {
			p.VCS.Branch = branch
			p.VCS.Dirty = p.VCS.Dirty || dirty
			p.VCS.Type = "git"
			return
		}
	}
	// Fallback to runtime current working directory
	if cwd, err := os.Getwd(); err == nil && cwd != "" {
		if branch, dirty := findGitBranch(cwd); branch != "" {
			p.VCS.Branch = branch
			p.VCS.Dirty = p.VCS.Dirty || dirty
			p.VCS.Type = "git"
		}
	}
}

// findGitBranch looks for a .git directory or worktree pointer starting from startDir.
func findGitBranch(startDir string) (string, bool) {
	curr := filepath.Clean(startDir)
	for {
		gitPath := filepath.Join(curr, ".git")
		fi, err := os.Stat(gitPath)
		if err == nil {
			if fi.IsDir() {
				headFile := filepath.Join(gitPath, "HEAD")
				if data, err := os.ReadFile(headFile); err == nil {
					return parseGitHead(string(data)), false
				}
			} else {
				// Worktree or submodule (.git file containing "gitdir: <path>")
				if data, err := os.ReadFile(gitPath); err == nil {
					line := strings.TrimSpace(string(data))
					if strings.HasPrefix(line, "gitdir:") {
						realGitDir := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
						if !filepath.IsAbs(realGitDir) {
							realGitDir = filepath.Join(curr, realGitDir)
						}
						headFile := filepath.Join(realGitDir, "HEAD")
						if hData, err := os.ReadFile(headFile); err == nil {
							return parseGitHead(string(hData)), false
						}
					}
				}
			}
			return "", false
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return "", false
}

// parseGitHead extracts the branch name or short commit SHA from HEAD file content.
func parseGitHead(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "ref: refs/heads/") {
		return strings.TrimPrefix(content, "ref: refs/heads/")
	}
	if len(content) >= 7 {
		return content[:7]
	}
	return content
}

// ActiveModelGroup identifies whether active model is first-party ("gemini") or third-party ("3p" / Claude / GPT).
func (p *StatePayload) ActiveModelGroup() string {
	if p == nil {
		return "gemini"
	}
	modelStr := strings.ToLower(p.Model.ID + " " + p.Model.DisplayName)
	if strings.Contains(modelStr, "claude") || strings.Contains(modelStr, "gpt") ||
		strings.Contains(modelStr, "opus") || strings.Contains(modelStr, "sonnet") ||
		strings.Contains(modelStr, "3p") || strings.Contains(modelStr, "anthropic") ||
		strings.Contains(modelStr, "openai") {
		return "3p"
	}
	return "gemini"
}

// GetWeeklyUsage returns the active weekly limit remaining percentage (0-100) and whether it was found.
func (p *StatePayload) GetWeeklyUsage() (float64, bool) {
	if p == nil {
		return 0, false
	}
	if p.WeeklyUsagePercent != nil {
		return min(100.0, max(0.0, *p.WeeklyUsagePercent)), true
	}
	if len(p.RateLimits) > 0 {
		for _, key := range []string{"seven_day", "7d", "weekly", "week"} {
			if rl, ok := p.RateLimits[key]; ok {
				if rl.RemainingPercentage > 0 {
					return min(100.0, max(0.0, rl.RemainingPercentage)), true
				}
				if rl.UsedPercentage > 0 {
					return min(100.0, max(0.0, 100.0-rl.UsedPercentage)), true
				}
			}
		}
	}
	if len(p.Quota) > 0 {
		group := p.ActiveModelGroup()
		var candidateKeys []string
		if group == "3p" {
			candidateKeys = []string{"3p-weekly", "3p_weekly", "3p-7d", "3p_7d", "claude-weekly", "claude_weekly", "gpt-weekly", "weekly", "7d"}
		} else {
			candidateKeys = []string{"gemini-weekly", "gemini_weekly", "gemini-7d", "gemini_7d", "weekly", "7d"}
		}

		for _, k := range candidateKeys {
			if q, ok := p.Quota[k]; ok && (q.RemainingFraction > 0 || q.UsedPercentage > 0) {
				if q.RemainingFraction > 0 {
					return min(100.0, max(0.0, q.RemainingFraction*100.0)), true
				}
				if q.UsedPercentage > 0 {
					return min(100.0, max(0.0, 100.0-q.UsedPercentage)), true
				}
			}
		}

		// Fuzzy search in quota keys
		for key, q := range p.Quota {
			k := strings.ToLower(key)
			if strings.Contains(k, "week") || strings.Contains(k, "7d") || strings.Contains(k, "seven") {
				if (group == "gemini" && strings.Contains(k, "gemini")) ||
					(group == "3p" && (strings.Contains(k, "3p") || strings.Contains(k, "claude") || strings.Contains(k, "gpt"))) {
					if q.RemainingFraction > 0 {
						return min(100.0, max(0.0, q.RemainingFraction*100.0)), true
					}
					if q.UsedPercentage > 0 {
						return min(100.0, max(0.0, 100.0-q.UsedPercentage)), true
					}
				}
			}
		}
	}
	return 0, false
}

// GetSessionUsage returns the active 5-hour limit remaining percentage (0-100) and whether it was found.
func (p *StatePayload) GetSessionUsage() (float64, bool) {
	if p == nil {
		return 0, false
	}
	if p.SessionUsagePercent != nil {
		return min(100.0, max(0.0, *p.SessionUsagePercent)), true
	}
	if len(p.RateLimits) > 0 {
		for _, key := range []string{"five_hour", "5h", "session", "current"} {
			if rl, ok := p.RateLimits[key]; ok {
				if rl.RemainingPercentage > 0 {
					return min(100.0, max(0.0, rl.RemainingPercentage)), true
				}
				if rl.UsedPercentage > 0 {
					return min(100.0, max(0.0, 100.0-rl.UsedPercentage)), true
				}
			}
		}
	}
	if len(p.Quota) > 0 {
		group := p.ActiveModelGroup()
		var candidateKeys []string
		if group == "3p" {
			candidateKeys = []string{"3p-5h", "3p_5h", "3p", "claude-5h", "claude_5h", "gpt-5h", "gpt_5h", "session", "5h"}
		} else {
			candidateKeys = []string{"gemini-5h", "gemini_5h", "gemini", "session", "5h"}
		}

		for _, k := range candidateKeys {
			if q, ok := p.Quota[k]; ok && (q.RemainingFraction > 0 || q.UsedPercentage > 0) {
				if q.RemainingFraction > 0 {
					return min(100.0, max(0.0, q.RemainingFraction*100.0)), true
				}
				if q.UsedPercentage > 0 {
					return min(100.0, max(0.0, 100.0-q.UsedPercentage)), true
				}
			}
		}

		// Fuzzy search in quota keys
		for key, q := range p.Quota {
			k := strings.ToLower(key)
			if strings.Contains(k, "week") || strings.Contains(k, "7d") || strings.Contains(k, "seven") {
				continue
			}
			if (group == "gemini" && (strings.Contains(k, "gemini") || strings.Contains(k, "flash") || strings.Contains(k, "pro"))) ||
				(group == "3p" && (strings.Contains(k, "3p") || strings.Contains(k, "claude") || strings.Contains(k, "gpt") || strings.Contains(k, "sonnet") || strings.Contains(k, "opus"))) {
				if q.RemainingFraction > 0 {
					return min(100.0, max(0.0, q.RemainingFraction*100.0)), true
				}
				if q.UsedPercentage > 0 {
					return min(100.0, max(0.0, 100.0-q.UsedPercentage)), true
				}
			}
		}
	}
	return 0, false
}
