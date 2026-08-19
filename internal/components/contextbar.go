package components

import (
	"strconv"
	"strings"

	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

func RenderContextBar(ctx payload.ContextWindowInfo, barConfig config.ContextBarConfig, p theme.Palette, icons config.IconsConfig) string {
	barLength := barConfig.Length
	if barLength <= 0 {
		barLength = 10
	}

	percentage := min(100.0, max(0.0, ctx.UsedPercentage))

	colorCode := selectContextColor(percentage, barConfig, p)
	barString := buildSmoothBar(percentage, barLength)

	var numBuf [16]byte
	numSlice := strconv.AppendFloat(numBuf[:0], percentage, 'f', 1, 64)
	numSlice = append(numSlice, '%')
	pctFormatted := string(numSlice)

	coloredBar := theme.Style(colorCode, barString)
	numHighlight := theme.Style(p.Highlight, pctFormatted)

	icon := theme.IconContext
	if icons.Context != "" {
		icon = icons.Context
	}

	var b strings.Builder
	b.Grow(len(icon) + 1 + len(coloredBar) + 1 + len(numHighlight))
	b.WriteString(icon)
	b.WriteByte(' ')
	b.WriteString(coloredBar)
	b.WriteByte(' ')
	b.WriteString(numHighlight)
	return b.String()
}

func selectContextColor(percentage float64, barConfig config.ContextBarConfig, p theme.Palette) string {
	critThreshold := barConfig.CriticalThreshold
	if critThreshold <= 0 {
		critThreshold = 90.0
	}
	warnThreshold := barConfig.WarningThreshold
	if warnThreshold <= 0 {
		warnThreshold = 75.0
	}

	if percentage >= critThreshold {
		return p.ContextCrit
	}
	if percentage >= warnThreshold {
		return p.ContextWarn
	}
	return p.ContextOK
}

func buildSmoothBar(percentage float64, totalSegments int) string {
	exactUnits := (percentage / 100.0) * float64(totalSegments)
	fullBlocks := int(exactUnits)
	remainder := exactUnits - float64(fullBlocks)

	var builder strings.Builder
	builder.Grow(totalSegments * 4)

	limit := min(fullBlocks, totalSegments)
	for i := range limit {
		_ = i
		builder.WriteString(theme.BlockFull)
	}

	if fullBlocks < totalSegments {
		builder.WriteString(selectSubBlock(remainder))
		for i := fullBlocks + 1; i < totalSegments; i++ {
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
