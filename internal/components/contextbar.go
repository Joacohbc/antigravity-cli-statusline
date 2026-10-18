package components

import (
	"fmt"
	"strings"

	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

func RenderContextBar(contextInfo payload.ContextWindowInfo, barConfig config.ContextBarConfig, palette theme.Palette, icons config.IconsConfig) string {
	barLength := barConfig.Length
	if barLength <= 0 {
		barLength = 10
	}

	percentage := min(100.0, max(0.0, contextInfo.UsedPercentage))

	colorCode := selectContextColor(percentage, barConfig, palette)
	barString := buildSmoothBar(percentage, barLength)

	percentFormatted := fmt.Sprintf("%.1f%%", percentage)
	coloredBar := theme.Style(colorCode, barString)
	numberHighlight := theme.Style(palette.Highlight, percentFormatted)

	icon := theme.IconContext
	if icons.Context != "" {
		icon = icons.Context
	}

	var builder strings.Builder
	builder.Grow(len(icon) + 1 + len(coloredBar) + 1 + len(numberHighlight) + 8)
	builder.WriteString(icon)
	builder.WriteByte(' ')
	builder.WriteString(coloredBar)
	builder.WriteByte(' ')
	builder.WriteString(numberHighlight)
	if percentage >= 85.0 {
		builder.WriteByte(' ')
		builder.WriteString(theme.Style(palette.Attention, "󰘚"))
	}
	return builder.String()
}

func selectContextColor(percentage float64, barConfig config.ContextBarConfig, palette theme.Palette) string {
	criticalThreshold := barConfig.CriticalThreshold
	if criticalThreshold <= 0 {
		criticalThreshold = 90.0
	}
	warningThreshold := barConfig.WarningThreshold
	if warningThreshold <= 0 {
		warningThreshold = 75.0
	}

	if percentage >= criticalThreshold {
		return palette.ContextCrit
	}
	if percentage >= warningThreshold {
		return palette.ContextWarn
	}
	return palette.ContextOK
}

func buildSmoothBar(percentage float64, totalSegments int) string {
	exactUnits := (percentage / 100.0) * float64(totalSegments)
	fullBlocks := int(exactUnits)
	remainder := exactUnits - float64(fullBlocks)

	var builder strings.Builder
	builder.Grow(totalSegments * 4)

	limit := min(fullBlocks, totalSegments)
	for index := range limit {
		_ = index
		builder.WriteString(theme.BlockFull)
	}

	if fullBlocks < totalSegments {
		builder.WriteString(selectSubBlock(remainder))
		for index := fullBlocks + 1; index < totalSegments; index++ {
			builder.WriteString(theme.BlockEmpty)
		}
	}

	return builder.String()
}

func selectSubBlock(remainder float64) string {
	switch {
	case remainder >= 0.875:
		return theme.Block78ths
	case remainder >= 0.75:
		return theme.Block34ths
	case remainder >= 0.625:
		return theme.Block58ths
	case remainder >= 0.50:
		return theme.BlockHalf
	case remainder >= 0.375:
		return theme.Block38ths
	case remainder >= 0.25:
		return theme.Block14th
	case remainder >= 0.125:
		return theme.Block18th
	default:
		return theme.BlockEmpty
	}
}
