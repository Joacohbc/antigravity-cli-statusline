package components

import (
	"cmp"
	"strings"

	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

func RenderModelBadge(model payload.ModelInfo, executionMode string, palette theme.Palette, icons config.IconsConfig) string {
	modelName := strings.TrimSpace(cmp.Or(payload.Sanitize(model.DisplayName), payload.Sanitize(model.ID)))
	if modelName == "" {
		return ""
	}

	icon := theme.IconModel
	if icons.Model != "" {
		icon = icons.Model
	}

	modelText := theme.Style(palette.Model, icon+" "+modelName)

	sanitizedMode := strings.ToUpper(strings.TrimSpace(payload.Sanitize(executionMode)))
	if sanitizedMode == "" || sanitizedMode == "DEFAULT" {
		return modelText
	}

	modeTag := theme.Style(palette.ModelMode, "["+sanitizedMode+"]")
	var builder strings.Builder
	builder.Grow(len(modelText) + 1 + len(modeTag))
	builder.WriteString(modelText)
	builder.WriteByte(' ')
	builder.WriteString(modeTag)
	return builder.String()
}
