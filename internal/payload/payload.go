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
	MaxPayloadSize       = 2 * 1024 * 1024
	DefaultTerminalWidth = 80
	DefaultAgentState    = "idle"
)

type ModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

func (model ModelInfo) GetID() string {
	return cmp.Or(Sanitize(model.ID), "unknown")
}

func (model ModelInfo) GetDisplayName() string {
	return cmp.Or(Sanitize(model.DisplayName), model.GetID())
}

func (model ModelInfo) Name() string {
	return model.GetDisplayName()
}

type WorkspaceInfo struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

func (workspace WorkspaceInfo) GetCurrentDir() string {
	return Sanitize(workspace.CurrentDir)
}

func (workspace WorkspaceInfo) GetProjectDir() string {
	return Sanitize(workspace.ProjectDir)
}

type TokenUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

func (usage TokenUsage) GetInputTokens() int {
	return max(0, usage.InputTokens)
}

func (usage TokenUsage) GetOutputTokens() int {
	return max(0, usage.OutputTokens)
}

func (usage TokenUsage) GetCacheCreationInputTokens() int {
	return max(0, usage.CacheCreationInputTokens)
}

func (usage TokenUsage) GetCacheReadInputTokens() int {
	return max(0, usage.CacheReadInputTokens)
}

func (usage TokenUsage) TotalTokens() int {
	return usage.GetInputTokens() + usage.GetOutputTokens()
}

type ContextWindowInfo struct {
	TotalInputTokens    int        `json:"total_input_tokens"`
	TotalOutputTokens   int        `json:"total_output_tokens"`
	ContextWindowSize   int        `json:"context_window_size"`
	UsedPercentage      float64    `json:"used_percentage"`
	RemainingPercentage float64    `json:"remaining_percentage"`
	CurrentUsage        TokenUsage `json:"current_usage"`
}

type ContextInfo = ContextWindowInfo

func (contextWindow ContextWindowInfo) GetTotalInputTokens() int {
	return max(0, contextWindow.TotalInputTokens)
}

func (contextWindow ContextWindowInfo) GetTotalOutputTokens() int {
	return max(0, contextWindow.TotalOutputTokens)
}

func (contextWindow ContextWindowInfo) GetContextWindowSize() int {
	return max(0, contextWindow.ContextWindowSize)
}

func (contextWindow ContextWindowInfo) GetUsedPercentage() float64 {
	return min(100.0, max(0.0, contextWindow.UsedPercentage))
}

func (contextWindow ContextWindowInfo) GetRemainingPercentage() float64 {
	if contextWindow.RemainingPercentage == 0.0 && contextWindow.UsedPercentage <= 0.0 {
		return 100.0
	}
	if contextWindow.RemainingPercentage == 0.0 && contextWindow.UsedPercentage > 0.0 {
		return max(0.0, 100.0-contextWindow.UsedPercentage)
	}
	return min(100.0, max(0.0, contextWindow.RemainingPercentage))
}

func (contextWindow ContextWindowInfo) GetCurrentUsage() TokenUsage {
	return contextWindow.CurrentUsage
}

type VCSInfo struct {
	Type   string `json:"type"`
	Branch string `json:"branch"`
	Client string `json:"client"`
	Dirty  bool   `json:"dirty"`
}

func (vcs VCSInfo) GetType() string {
	return cmp.Or(Sanitize(vcs.Type), "git")
}

func (vcs VCSInfo) GetBranch() string {
	return Sanitize(vcs.Branch)
}

func (vcs VCSInfo) GetClient() string {
	return Sanitize(vcs.Client)
}

func (vcs VCSInfo) IsDirty() bool {
	return vcs.Dirty
}

type SandboxInfo struct {
	Enabled      bool `json:"enabled"`
	AllowNetwork bool `json:"allow_network"`
}

func (sandbox SandboxInfo) IsEnabled() bool {
	return sandbox.Enabled
}

func (sandbox SandboxInfo) IsAllowNetwork() bool {
	return sandbox.AllowNetwork
}

type QuotaDetail struct {
	RemainingFraction float64 `json:"remaining_fraction"`
	UsedPercentage    float64 `json:"used_percentage,omitempty"`
	ResetTime         string  `json:"reset_time"`
	ResetInSeconds    int     `json:"reset_in_seconds"`
}

type RateLimit struct {
	UsedPercentage      float64 `json:"used_percentage"`
	RemainingPercentage float64 `json:"remaining_percentage"`
	ResetTime           string  `json:"reset_time"`
	ResetInSeconds      int     `json:"reset_in_seconds"`
}

func (quota QuotaDetail) GetRemainingFraction() float64 {
	if quota.RemainingFraction == 0.0 && quota.ResetTime == "" && quota.ResetInSeconds == 0 {
		return 1.0
	}
	return min(1.0, max(0.0, quota.RemainingFraction))
}

func (quota QuotaDetail) GetResetTime() string {
	return Sanitize(quota.ResetTime)
}

func (quota QuotaDetail) GetResetInSeconds() int {
	return max(0, quota.ResetInSeconds)
}

type QuotaInfo = map[string]QuotaDetail

type VimInfo struct {
	Mode string `json:"mode"`
}

func (vim *VimInfo) GetMode() string {
	if vim == nil {
		return ""
	}
	return Sanitize(vim.Mode)
}

type SubagentDetail struct {
	Role  string `json:"role"`
	Type  string `json:"type"`
	State string `json:"state"`
}

type StatsInfo struct {
	ArtifactCount int `json:"artifact_count"`
	TaskCount     int `json:"task_count"`
	SubagentCount int `json:"subagent_count"`
	TerminalCount int `json:"terminal_count"`
	MCPCount      int `json:"mcp_count"`
	SkillsCount   int `json:"skills_count"`
	PluginsCount  int `json:"plugins_count"`
}

func (stats StatsInfo) GetArtifactCount() int {
	return max(0, stats.ArtifactCount)
}

func (stats StatsInfo) GetTaskCount() int {
	return max(0, stats.TaskCount)
}

func (stats StatsInfo) GetSubagentCount() int {
	return max(0, stats.SubagentCount)
}

func (stats StatsInfo) GetTerminalCount() int {
	return max(0, stats.TerminalCount)
}

func (stats StatsInfo) GetMCPCount() int {
	return max(0, stats.MCPCount)
}

func (stats StatsInfo) GetSkillsCount() int {
	return max(0, stats.SkillsCount)
}

func (stats StatsInfo) GetPluginsCount() int {
	return max(0, stats.PluginsCount)
}

type StatePayload struct {
	CWD                     string                     `json:"cwd"`
	SessionID               string                     `json:"session_id"`
	ConversationID          string                     `json:"conversation_id"`
	TranscriptPath          string                     `json:"transcript_path"`
	Model                   ModelInfo                  `json:"model"`
	Workspace               WorkspaceInfo              `json:"workspace"`
	Version                 string                     `json:"version"`
	ContextWindow           ContextWindowInfo          `json:"context_window"`
	Exceeds200kTokens       *bool                      `json:"exceeds_200k_tokens"`
	Product                 string                     `json:"product"`
	Quota                   map[string]QuotaDetail     `json:"quota"`
	WeeklyUsagePercent      *float64                   `json:"weekly_usage_percent,omitempty"`
	SessionUsagePercent     *float64                   `json:"session_usage_percent,omitempty"`
	RateLimits              map[string]RateLimit       `json:"rate_limits,omitempty"`
	AgentState              string                     `json:"agent_state"`
	VCS                     VCSInfo                    `json:"vcs"`
	Sandbox                 SandboxInfo                `json:"sandbox"`
	ArtifactCount           int                        `json:"artifact_count"`
	PlanTier                string                     `json:"plan_tier"`
	Email                   string                     `json:"email"`
	PendingInputCount       int                        `json:"pending_input_count"`
	ToolConfirmationPending bool                       `json:"tool_confirmation_pending"`
	TaskCount               int                        `json:"task_count"`
	SubagentCount           int                        `json:"-"`
	RawSubagents            json.RawMessage            `json:"subagents"`
	SubagentDetails         []SubagentDetail           `json:"-"`
	TerminalCount           int                        `json:"terminal_count"`
	MCPCount                int                        `json:"mcp_count"`
	SkillsCount             int                        `json:"skills_count"`
	PluginsCount            int                        `json:"plugins_count"`
	TerminalWidth           int                        `json:"terminal_width"`
	ExecutionMode           string                     `json:"execution_mode"`
	Vim                     *VimInfo                   `json:"vim"`
	RawMap                  map[string]json.RawMessage `json:"-"`
}

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

func (payload *StatePayload) GetCWD() string {
	if payload == nil {
		return ""
	}
	return Sanitize(payload.CWD)
}

func (payload *StatePayload) GetSessionID() string {
	if payload == nil {
		return ""
	}
	return Sanitize(payload.SessionID)
}

func (payload *StatePayload) GetConversationID() string {
	if payload == nil {
		return ""
	}
	return Sanitize(payload.ConversationID)
}

func (payload *StatePayload) GetTranscriptPath() string {
	if payload == nil {
		return ""
	}
	return Sanitize(payload.TranscriptPath)
}

func (payload *StatePayload) GetModel() ModelInfo {
	if payload == nil {
		return ModelInfo{ID: "unknown", DisplayName: "Unknown"}
	}
	return payload.Model
}

func (payload *StatePayload) GetWorkspace() WorkspaceInfo {
	if payload == nil {
		return WorkspaceInfo{}
	}
	return payload.Workspace
}

func (payload *StatePayload) GetVersion() string {
	if payload == nil {
		return ""
	}
	return Sanitize(payload.Version)
}

func (payload *StatePayload) GetContextWindow() ContextWindowInfo {
	if payload == nil {
		return ContextWindowInfo{}
	}
	return payload.ContextWindow
}

func (payload *StatePayload) GetProduct() string {
	if payload == nil {
		return ""
	}
	return Sanitize(payload.Product)
}

func (payload *StatePayload) GetQuota() map[string]QuotaDetail {
	if payload == nil || payload.Quota == nil {
		return make(map[string]QuotaDetail)
	}
	return maps.Clone(payload.Quota)
}

func (payload *StatePayload) GetAgentState() string {
	if payload == nil {
		return DefaultAgentState
	}
	return cmp.Or(Sanitize(payload.AgentState), DefaultAgentState)
}

func (payload *StatePayload) GetVCS() VCSInfo {
	if payload == nil {
		return VCSInfo{Type: "git"}
	}
	return payload.VCS
}

func (payload *StatePayload) GetSandbox() SandboxInfo {
	if payload == nil {
		return SandboxInfo{}
	}
	return payload.Sandbox
}

func (payload *StatePayload) GetArtifactCount() int {
	if payload == nil {
		return 0
	}
	return max(0, payload.ArtifactCount)
}

func (payload *StatePayload) GetPlanTier() string {
	if payload == nil {
		return ""
	}
	return Sanitize(payload.PlanTier)
}

func (payload *StatePayload) GetEmail() string {
	if payload == nil {
		return ""
	}
	return Sanitize(payload.Email)
}

func (payload *StatePayload) GetPendingInputCount() int {
	if payload == nil {
		return 0
	}
	return max(0, payload.PendingInputCount)
}

func (payload *StatePayload) IsToolConfirmationPending() bool {
	if payload == nil {
		return false
	}
	return payload.ToolConfirmationPending
}

func (payload *StatePayload) GetTaskCount() int {
	if payload == nil {
		return 0
	}
	return max(0, payload.TaskCount)
}

func (payload *StatePayload) GetSubagentCount() int {
	if payload == nil {
		return 0
	}
	return max(0, payload.SubagentCount)
}

func (payload *StatePayload) GetTerminalWidth() int {
	if payload == nil || payload.TerminalWidth <= 0 {
		return DefaultTerminalWidth
	}
	return payload.TerminalWidth
}

func (payload *StatePayload) GetExecutionMode() string {
	if payload == nil {
		return ""
	}
	return Sanitize(payload.ExecutionMode)
}

func (payload *StatePayload) GetVim() *VimInfo {
	if payload == nil {
		return nil
	}
	return payload.Vim
}

func (payload *StatePayload) GetTerminalCount() int {
	if payload == nil {
		return 0
	}
	return max(0, payload.TerminalCount)
}

func (payload *StatePayload) GetMCPCount() int {
	if payload == nil {
		return 0
	}
	return max(0, payload.MCPCount)
}

func (payload *StatePayload) GetSkillsCount() int {
	if payload == nil {
		return 0
	}
	return max(0, payload.SkillsCount)
}

func (payload *StatePayload) GetPluginsCount() int {
	if payload == nil {
		return 0
	}
	return max(0, payload.PluginsCount)
}

func (payload *StatePayload) GetSubagentDetails() []SubagentDetail {
	if payload == nil {
		return nil
	}
	return payload.SubagentDetails
}

func (payload *StatePayload) GetWaitingSubagentsCount() int {
	if payload == nil {
		return 0
	}
	waitingCount := 0
	for _, subagent := range payload.SubagentDetails {
		state := strings.ToLower(subagent.State)
		if strings.HasPrefix(state, "waiting") {
			waitingCount++
		}
	}
	return waitingCount
}

func (payload *StatePayload) GetStats() StatsInfo {
	if payload == nil {
		return StatsInfo{}
	}
	return StatsInfo{
		ArtifactCount: payload.GetArtifactCount(),
		TaskCount:     payload.GetTaskCount(),
		SubagentCount: payload.GetSubagentCount(),
		TerminalCount: payload.GetTerminalCount(),
		MCPCount:      payload.GetMCPCount(),
		SkillsCount:   payload.GetSkillsCount(),
		PluginsCount:  payload.GetPluginsCount(),
	}
}

func (payload *StatePayload) SanitizeFields() {
	if payload == nil {
		return
	}
	payload.CWD = Sanitize(payload.CWD)
	payload.SessionID = Sanitize(payload.SessionID)
	payload.ConversationID = Sanitize(payload.ConversationID)
	payload.TranscriptPath = Sanitize(payload.TranscriptPath)
	payload.Model.ID = Sanitize(payload.Model.ID)
	payload.Model.DisplayName = Sanitize(payload.Model.DisplayName)
	payload.Workspace.CurrentDir = Sanitize(payload.Workspace.CurrentDir)
	payload.Workspace.ProjectDir = Sanitize(payload.Workspace.ProjectDir)
	payload.Version = Sanitize(payload.Version)
	payload.Product = Sanitize(payload.Product)
	payload.AgentState = Sanitize(payload.AgentState)
	payload.VCS.Type = Sanitize(payload.VCS.Type)
	payload.VCS.Branch = Sanitize(payload.VCS.Branch)
	payload.VCS.Client = Sanitize(payload.VCS.Client)
	payload.PlanTier = Sanitize(payload.PlanTier)
	payload.Email = Sanitize(payload.Email)
	payload.ExecutionMode = Sanitize(payload.ExecutionMode)
	if payload.Vim != nil {
		payload.Vim.Mode = Sanitize(payload.Vim.Mode)
	}
}

func (payload *StatePayload) resolveSubagentCount() {
	if payload == nil || len(payload.RawSubagents) == 0 {
		return
	}

	var count int
	if err := json.Unmarshal(payload.RawSubagents, &count); err == nil {
		payload.SubagentCount = max(0, count)
		return
	}

	var details []SubagentDetail
	if err := json.Unmarshal(payload.RawSubagents, &details); err == nil {
		payload.SubagentDetails = details
		payload.SubagentCount = len(details)
		return
	}

	var list []any
	if err := json.Unmarshal(payload.RawSubagents, &list); err == nil {
		payload.SubagentCount = len(list)
	}
}

func (payload *StatePayload) resolveAdditionalCounts() {
	if payload == nil || payload.RawMap == nil {
		return
	}
	if payload.TerminalCount == 0 {
		if raw, ok := payload.RawMap["terminals"]; ok {
			var list []any
			if json.Unmarshal(raw, &list) == nil {
				payload.TerminalCount = len(list)
			}
		}
	}
	if payload.MCPCount == 0 {
		if raw, ok := payload.RawMap["mcp_servers"]; ok {
			var list []any
			if json.Unmarshal(raw, &list) == nil {
				payload.MCPCount = len(list)
			} else {
				var dict map[string]any
				if json.Unmarshal(raw, &dict) == nil {
					payload.MCPCount = len(dict)
				}
			}
		}
	}
	if payload.SkillsCount == 0 {
		if raw, ok := payload.RawMap["skills"]; ok {
			var list []any
			if json.Unmarshal(raw, &list) == nil {
				payload.SkillsCount = len(list)
			}
		}
	}
	if payload.PluginsCount == 0 {
		if raw, ok := payload.RawMap["plugins"]; ok {
			var list []any
			if json.Unmarshal(raw, &list) == nil {
				payload.PluginsCount = len(list)
			}
		}
	}
}

func Parse(reader io.Reader) (*StatePayload, error) {
	if reader == nil {
		return NewDefaultStatePayload(), &ParseError{
			Op:  "read",
			Err: ErrEmptyPayload,
			Msg: "reader is nil",
		}
	}

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
	payload.resolveAdditionalCounts()

	return payload, nil
}

func (payload *StatePayload) resolveVCS() {
	if payload.VCS.Branch != "" {
		return
	}

	directories := []string{payload.CWD, payload.Workspace.CurrentDir, payload.Workspace.ProjectDir}
	for _, directory := range directories {
		if directory == "" {
			continue
		}
		if branch, dirty := findGitBranch(directory); branch != "" {
			payload.VCS.Branch = branch
			payload.VCS.Dirty = payload.VCS.Dirty || dirty
			payload.VCS.Type = "git"
			return
		}
	}

	currentWorkingDirectory, err := os.Getwd()
	if err == nil && currentWorkingDirectory != "" {
		if branch, dirty := findGitBranch(currentWorkingDirectory); branch != "" {
			payload.VCS.Branch = branch
			payload.VCS.Dirty = payload.VCS.Dirty || dirty
			payload.VCS.Type = "git"
		}
	}
}

func readBranchFromDirectory(gitDirectory string) string {
	headFilePath := filepath.Join(gitDirectory, "HEAD")
	data, err := os.ReadFile(headFilePath)
	if err != nil {
		return ""
	}
	return parseGitHead(string(data))
}

func readBranchFromWorktree(gitFilePath, currentDirectory string) string {
	data, err := os.ReadFile(gitFilePath)
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, "gitdir:") {
		return ""
	}
	targetDirectory := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
	if !filepath.IsAbs(targetDirectory) {
		targetDirectory = filepath.Join(currentDirectory, targetDirectory)
	}
	return readBranchFromDirectory(targetDirectory)
}

func inspectGitPath(gitPath, currentDirectory string) (string, bool) {
	fileInfo, err := os.Stat(gitPath)
	if err != nil {
		return "", false
	}
	if fileInfo.IsDir() {
		branch := readBranchFromDirectory(gitPath)
		return branch, branch != ""
	}
	branch := readBranchFromWorktree(gitPath, currentDirectory)
	return branch, branch != ""
}

func findGitBranch(startDirectory string) (string, bool) {
	currentDirectory := filepath.Clean(startDirectory)
	for {
		gitPath := filepath.Join(currentDirectory, ".git")
		if branch, found := inspectGitPath(gitPath, currentDirectory); found {
			return branch, false
		}

		parentDirectory := filepath.Dir(currentDirectory)
		if parentDirectory == currentDirectory {
			return "", false
		}
		currentDirectory = parentDirectory
	}
}

func parseGitHead(content string) string {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "ref: refs/heads/") {
		return strings.TrimPrefix(trimmed, "ref: refs/heads/")
	}
	if len(trimmed) >= 7 {
		return trimmed[:7]
	}
	return trimmed
}

func (payload *StatePayload) ActiveModelGroup() string {
	if payload == nil {
		return "gemini"
	}
	modelText := strings.ToLower(payload.Model.ID + " " + payload.Model.DisplayName)
	isThirdParty := strings.Contains(modelText, "claude") || strings.Contains(modelText, "gpt") ||
		strings.Contains(modelText, "opus") || strings.Contains(modelText, "sonnet") ||
		strings.Contains(modelText, "3p") || strings.Contains(modelText, "anthropic") ||
		strings.Contains(modelText, "openai")
	if isThirdParty {
		return "3p"
	}
	return "gemini"
}

func resolveRateLimitUsage(rateLimits map[string]RateLimit, searchKeys []string) (float64, bool) {
	for _, searchKey := range searchKeys {
		rateLimit, exists := rateLimits[searchKey]
		if !exists {
			continue
		}
		if rateLimit.RemainingPercentage > 0 {
			return min(100.0, max(0.0, rateLimit.RemainingPercentage)), true
		}
		if rateLimit.UsedPercentage > 0 {
			return min(100.0, max(0.0, 100.0-rateLimit.UsedPercentage)), true
		}
	}
	return 0, false
}

func extractQuotaPercentage(quota QuotaDetail) (float64, bool) {
	if quota.RemainingFraction > 0 {
		return min(100.0, max(0.0, quota.RemainingFraction*100.0)), true
	}
	if quota.UsedPercentage > 0 {
		return min(100.0, max(0.0, 100.0-quota.UsedPercentage)), true
	}
	return 0, false
}

func resolveCandidateQuotaUsage(quotaMap map[string]QuotaDetail, candidateKeys []string) (float64, bool) {
	for _, candidateKey := range candidateKeys {
		quota, exists := quotaMap[candidateKey]
		if !exists {
			continue
		}
		if percentage, found := extractQuotaPercentage(quota); found {
			return percentage, true
		}
	}
	return 0, false
}

func matchesModelGroup(key, group string) bool {
	if group == "gemini" {
		return strings.Contains(key, "gemini") || strings.Contains(key, "flash") || strings.Contains(key, "pro")
	}
	return strings.Contains(key, "3p") || strings.Contains(key, "claude") || strings.Contains(key, "gpt") ||
		strings.Contains(key, "sonnet") || strings.Contains(key, "opus")
}

func resolveFuzzyQuotaUsage(quotaMap map[string]QuotaDetail, group string, isMatchingTimeWindow func(string) bool) (float64, bool) {
	for key, quota := range quotaMap {
		keyLower := strings.ToLower(key)
		if !isMatchingTimeWindow(keyLower) || !matchesModelGroup(keyLower, group) {
			continue
		}
		if percentage, found := extractQuotaPercentage(quota); found {
			return percentage, true
		}
	}
	return 0, false
}

func (payload *StatePayload) GetWeeklyUsage() (float64, bool) {
	if payload == nil {
		return 0, false
	}
	if payload.WeeklyUsagePercent != nil {
		return min(100.0, max(0.0, *payload.WeeklyUsagePercent)), true
	}
	if percentage, found := resolveRateLimitUsage(payload.RateLimits, []string{"seven_day", "7d", "weekly", "week"}); found {
		return percentage, true
	}
	if len(payload.Quota) == 0 {
		return 0, false
	}

	group := payload.ActiveModelGroup()
	candidateKeys := []string{"gemini-weekly", "gemini_weekly", "gemini-7d", "gemini_7d", "weekly", "7d"}
	if group == "3p" {
		candidateKeys = []string{"3p-weekly", "3p_weekly", "3p-7d", "3p_7d", "claude-weekly", "claude_weekly", "gpt-weekly", "weekly", "7d"}
	}
	if percentage, found := resolveCandidateQuotaUsage(payload.Quota, candidateKeys); found {
		return percentage, true
	}

	return resolveFuzzyQuotaUsage(payload.Quota, group, func(key string) bool {
		return strings.Contains(key, "week") || strings.Contains(key, "7d") || strings.Contains(key, "seven")
	})
}

func (payload *StatePayload) GetSessionUsage() (float64, bool) {
	if payload == nil {
		return 0, false
	}
	if payload.SessionUsagePercent != nil {
		return min(100.0, max(0.0, *payload.SessionUsagePercent)), true
	}
	if percentage, found := resolveRateLimitUsage(payload.RateLimits, []string{"five_hour", "5h", "session", "current"}); found {
		return percentage, true
	}
	if len(payload.Quota) == 0 {
		return 0, false
	}

	group := payload.ActiveModelGroup()
	candidateKeys := []string{"gemini-5h", "gemini_5h", "gemini", "session", "5h"}
	if group == "3p" {
		candidateKeys = []string{"3p-5h", "3p_5h", "3p", "claude-5h", "claude_5h", "gpt-5h", "gpt_5h", "session", "5h"}
	}
	if percentage, found := resolveCandidateQuotaUsage(payload.Quota, candidateKeys); found {
		return percentage, true
	}

	return resolveFuzzyQuotaUsage(payload.Quota, group, func(key string) bool {
		hasWeek := strings.Contains(key, "week") || strings.Contains(key, "7d") || strings.Contains(key, "seven")
		return !hasWeek
	})
}
