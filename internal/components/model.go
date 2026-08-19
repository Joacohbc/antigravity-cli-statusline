package components

import (
	"cmp"
	"strings"

	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

func RenderModelBadge(model payload.ModelInfo, executionMode string, p theme.Palette, icons config.IconsConfig) string {
	name := strings.TrimSpace(cmp.Or(payload.Sanitize(model.DisplayName), payload.Sanitize(model.ID)))
	if name == "" {
		return ""
	}

	icon := theme.IconModel
	if icons.Model != "" {
		icon = icons.Model
	}

	modelText := theme.Style(p.Model, icon+" "+name)

	mode := strings.ToUpper(strings.TrimSpace(payload.Sanitize(executionMode)))
	if mode == "" || mode == "DEFAULT" {
		return modelText
	}

	modeTag := theme.Style(p.ModelMode, "["+mode+"]")
	var b strings.Builder
	b.Grow(len(modelText) + 1 + len(modeTag))
	b.WriteString(modelText)
	b.WriteByte(' ')
	b.WriteString(modeTag)
	return b.String()
}
