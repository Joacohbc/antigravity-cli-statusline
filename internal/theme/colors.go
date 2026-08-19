package theme

import (
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"

	FgGray          = "\033[90m"
	FgBrightRed     = "\033[91m"
	FgBrightGreen   = "\033[92m"
	FgBrightYellow  = "\033[93m"
	FgBrightBlue    = "\033[94m"
	FgBrightMagenta = "\033[95m"
	FgBrightCyan    = "\033[96m"
	FgBrightWhite   = "\033[97m"

	BgBlack         = "\033[40m"
	BgRed           = "\033[41m"
	BgGreen         = "\033[42m"
	BgYellow        = "\033[43m"
	BgBlue          = "\033[44m"
	BgMagenta       = "\033[45m"
	BgCyan          = "\033[46m"
	BgWhite         = "\033[47m"
	BgDarkGray      = "\033[100m"
	BgBrightRed     = "\033[101m"
	BgBrightGreen   = "\033[102m"
	BgBrightYellow  = "\033[103m"
	BgBrightBlue    = "\033[104m"
	BgBrightMagenta = "\033[105m"
	BgBrightCyan    = "\033[106m"
	BgBrightWhite   = "\033[107m"
)

const (
	colorAuto         int32 = 0
	colorForceEnabled int32 = 1
	colorForceDisable int32 = 2
)

var colorOverride atomic.Int32

func SetColorEnabled(enabled bool) {
	if enabled {
		colorOverride.Store(colorForceEnabled)
	} else {
		colorOverride.Store(colorForceDisable)
	}
}

func ResetColorOverride() {
	colorOverride.Store(colorAuto)
}

func IsNoColor() bool {
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		if os.Getenv("NO_COLOR") != "" {
			return true
		}
	}
	if os.Getenv("TERM") == "dumb" {
		return true
	}
	return false
}

func ColorEnabled() bool {
	switch colorOverride.Load() {
	case colorForceEnabled:
		return true
	case colorForceDisable:
		return false
	default:
		return !IsNoColor()
	}
}

func Fg256(code int) string {
	return "\033[38;5;" + strconv.Itoa(code) + "m"
}

func Bg256(code int) string {
	return "\033[48;5;" + strconv.Itoa(code) + "m"
}

func FgRGB(r, g, b uint8) string {
	return "\033[38;2;" + strconv.Itoa(int(r)) + ";" + strconv.Itoa(int(g)) + ";" + strconv.Itoa(int(b)) + "m"
}

func BgRGB(r, g, b uint8) string {
	return "\033[48;2;" + strconv.Itoa(int(r)) + ";" + strconv.Itoa(int(g)) + ";" + strconv.Itoa(int(b)) + "m"
}

func FgTrueColor(r, g, b uint8) string {
	return FgRGB(r, g, b)
}

func BgTrueColor(r, g, b uint8) string {
	return BgRGB(r, g, b)
}

func Style(styleCode string, text string) string {
	if text == "" {
		return ""
	}
	if !ColorEnabled() || styleCode == "" {
		return text
	}
	var b strings.Builder
	b.Grow(len(styleCode) + len(text) + len(Reset))
	b.WriteString(styleCode)
	b.WriteString(text)
	b.WriteString(Reset)
	return b.String()
}

func Paint(fgCode string, text string) string {
	return Style(fgCode, text)
}

func Capsule(fgColor string, bgColor string, capFgColor string, text string) string {
	if text == "" {
		return ""
	}
	if !ColorEnabled() {
		return " " + text + " "
	}
	leftCap := Style(capFgColor, PillLeft)
	body := Style(bgColor+fgColor+Bold, " "+text+" ")
	rightCap := Style(capFgColor, PillRight)

	var b strings.Builder
	b.Grow(len(leftCap) + len(body) + len(rightCap))
	b.WriteString(leftCap)
	b.WriteString(body)
	b.WriteString(rightCap)
	return b.String()
}

func Clean(s string) string {
	if s == "" {
		return ""
	}

	hasEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b || (s[i] < 32 && s[i] != '\t' && s[i] != '\n') {
			hasEscape = true
			break
		}
	}
	if !hasEscape {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	i := 0
	n := len(s)
	for i < n {
		if s[i] == 0x1b {
			i++
			if i >= n {
				break
			}
			if s[i] == '[' {
				i++
				for i < n && !((s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') || s[i] == '~') {
					i++
				}
				if i < n {
					i++
				}
			} else if s[i] == ']' {
				i++
				for i < n {
					if s[i] == 0x07 {
						i++
						break
					}
					if s[i] == 0x1b && i+1 < n && s[i+1] == '\\' {
						i += 2
						break
					}
					i++
				}
			} else if s[i] == '(' || s[i] == ')' {
				i += 2
			} else {
				i++
			}
		} else if s[i] < 32 && s[i] != '\t' && s[i] != '\n' {
			i++
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}

func RuneWidth(r rune) int {
	if r < 32 || (r >= 0x7f && r < 0xa0) {
		return 0
	}
	if (r >= 0x0300 && r <= 0x036F) ||
		(r >= 0x1AB0 && r <= 0x1AFF) ||
		(r >= 0x1DC0 && r <= 0x1DFF) ||
		(r >= 0x200B && r <= 0x200F) ||
		(r >= 0x202A && r <= 0x202E) ||
		(r >= 0x2060 && r <= 0x206F) ||
		(r >= 0xFE00 && r <= 0xFE0F) ||
		(r >= 0xE0100 && r <= 0xE01EF) {
		return 0
	}
	if (r >= 0x1100 && r <= 0x115F) ||
		(r >= 0x2329 && r <= 0x232A) ||
		(r >= 0x2E80 && r <= 0xA4CF && r != 0x303F) ||
		(r >= 0xAC00 && r <= 0xD7A3) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE10 && r <= 0xFE19) ||
		(r >= 0xFE30 && r <= 0xFE6F) ||
		(r >= 0xFF00 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x1F300 && r <= 0x1F64F) ||
		(r >= 0x1F680 && r <= 0x1F6FF) ||
		(r >= 0x1F900 && r <= 0x1F9FF) ||
		(r >= 0x1FA70 && r <= 0x1FAFF) ||
		(r >= 0x20000 && r <= 0x3FFFD) {
		return 2
	}
	return 1
}

func VisibleWidth(s string) int {
	cleaned := Clean(s)
	width := 0
	for _, r := range cleaned {
		width += RuneWidth(r)
	}
	return width
}

type Palette struct {
	Name        string
	Ready       string
	Thinking    string
	Working     string
	Tool        string
	Init        string
	Unknown     string
	Branch      string
	BranchDirty string
	DirtyMarker string
	Model       string
	ModelMode   string
	ContextOK   string
	ContextWarn string
	ContextCrit string
	Artifact    string
	Subagent    string
	Task        string
	SandboxON   string
	SandboxOFF  string
	VimNormal   string
	VimInsert   string
	VimVisual   string
	Border      string
	Separator   string
	Dim         string
	Highlight   string
}

func GetPalette(name string) Palette {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "tokyonight":
		return Palette{
			Name:        "tokyonight",
			Ready:       FgRGB(158, 206, 106) + Bold,
			Thinking:    FgRGB(224, 175, 104) + Bold,
			Working:     FgRGB(125, 207, 255) + Bold,
			Tool:        FgRGB(187, 154, 247) + Bold,
			Init:        FgRGB(122, 162, 247) + Bold,
			Unknown:     FgRGB(192, 202, 245) + Bold,
			Branch:      FgRGB(122, 162, 247) + Bold,
			BranchDirty: FgRGB(247, 118, 142) + Bold,
			DirtyMarker: FgRGB(224, 175, 104) + Bold,
			Model:       FgRGB(187, 154, 247) + Italic,
			ModelMode:   FgRGB(224, 175, 104) + Bold,
			ContextOK:   FgRGB(192, 202, 245),
			ContextWarn: FgRGB(224, 175, 104),
			ContextCrit: FgRGB(247, 118, 142),
			Artifact:    FgRGB(192, 202, 245),
			Subagent:    FgRGB(125, 207, 255) + Bold,
			Task:        FgRGB(192, 202, 245),
			SandboxON:   FgRGB(158, 206, 106) + Bold,
			SandboxOFF:  FgRGB(86, 95, 137),
			VimNormal:   FgRGB(122, 162, 247) + Bold,
			VimInsert:   FgRGB(158, 206, 106) + Bold,
			VimVisual:   FgRGB(224, 175, 104) + Bold,
			Border:      FgRGB(86, 95, 137),
			Separator:   FgRGB(86, 95, 137),
			Dim:         FgRGB(86, 95, 137),
			Highlight:   FgRGB(255, 255, 255) + Bold,
		}
	case "catppuccin":
		return Palette{
			Name:        "catppuccin",
			Ready:       FgRGB(166, 227, 161) + Bold,
			Thinking:    FgRGB(249, 226, 175) + Bold,
			Working:     FgRGB(137, 220, 235) + Bold,
			Tool:        FgRGB(203, 166, 247) + Bold,
			Init:        FgRGB(137, 180, 250) + Bold,
			Unknown:     FgRGB(205, 214, 244) + Bold,
			Branch:      FgRGB(137, 180, 250) + Bold,
			BranchDirty: FgRGB(243, 139, 168) + Bold,
			DirtyMarker: FgRGB(249, 226, 175) + Bold,
			Model:       FgRGB(203, 166, 247) + Italic,
			ModelMode:   FgRGB(249, 226, 175) + Bold,
			ContextOK:   FgRGB(205, 214, 244),
			ContextWarn: FgRGB(249, 226, 175),
			ContextCrit: FgRGB(243, 139, 168),
			Artifact:    FgRGB(205, 214, 244),
			Subagent:    FgRGB(137, 220, 235) + Bold,
			Task:        FgRGB(205, 214, 244),
			SandboxON:   FgRGB(166, 227, 161) + Bold,
			SandboxOFF:  FgRGB(108, 112, 134),
			VimNormal:   FgRGB(137, 180, 250) + Bold,
			VimInsert:   FgRGB(166, 227, 161) + Bold,
			VimVisual:   FgRGB(249, 226, 175) + Bold,
			Border:      FgRGB(108, 112, 134),
			Separator:   FgRGB(108, 112, 134),
			Dim:         FgRGB(108, 112, 134),
			Highlight:   FgRGB(255, 255, 255) + Bold,
		}
	case "dracula":
		return Palette{
			Name:        "dracula",
			Ready:       FgRGB(80, 250, 123) + Bold,
			Thinking:    FgRGB(241, 250, 140) + Bold,
			Working:     FgRGB(139, 233, 253) + Bold,
			Tool:        FgRGB(255, 121, 198) + Bold,
			Init:        FgRGB(189, 147, 249) + Bold,
			Unknown:     FgRGB(248, 248, 242) + Bold,
			Branch:      FgRGB(189, 147, 249) + Bold,
			BranchDirty: FgRGB(255, 85, 85) + Bold,
			DirtyMarker: FgRGB(241, 250, 140) + Bold,
			Model:       FgRGB(255, 121, 198) + Italic,
			ModelMode:   FgRGB(241, 250, 140) + Bold,
			ContextOK:   FgRGB(248, 248, 242),
			ContextWarn: FgRGB(241, 250, 140),
			ContextCrit: FgRGB(255, 85, 85),
			Artifact:    FgRGB(248, 248, 242),
			Subagent:    FgRGB(139, 233, 253) + Bold,
			Task:        FgRGB(248, 248, 242),
			SandboxON:   FgRGB(80, 250, 123) + Bold,
			SandboxOFF:  FgRGB(98, 114, 164),
			VimNormal:   FgRGB(189, 147, 249) + Bold,
			VimInsert:   FgRGB(80, 250, 123) + Bold,
			VimVisual:   FgRGB(241, 250, 140) + Bold,
			Border:      FgRGB(98, 114, 164),
			Separator:   FgRGB(98, 114, 164),
			Dim:         FgRGB(98, 114, 164),
			Highlight:   FgRGB(255, 255, 255) + Bold,
		}
	case "nord":
		return Palette{
			Name:        "nord",
			Ready:       FgRGB(163, 190, 140) + Bold,
			Thinking:    FgRGB(235, 203, 139) + Bold,
			Working:     FgRGB(136, 192, 208) + Bold,
			Tool:        FgRGB(180, 142, 173) + Bold,
			Init:        FgRGB(129, 161, 193) + Bold,
			Unknown:     FgRGB(236, 239, 244) + Bold,
			Branch:      FgRGB(129, 161, 193) + Bold,
			BranchDirty: FgRGB(191, 97, 106) + Bold,
			DirtyMarker: FgRGB(235, 203, 139) + Bold,
			Model:       FgRGB(180, 142, 173) + Italic,
			ModelMode:   FgRGB(235, 203, 139) + Bold,
			ContextOK:   FgRGB(236, 239, 244),
			ContextWarn: FgRGB(235, 203, 139),
			ContextCrit: FgRGB(191, 97, 106),
			Artifact:    FgRGB(236, 239, 244),
			Subagent:    FgRGB(136, 192, 208) + Bold,
			Task:        FgRGB(236, 239, 244),
			SandboxON:   FgRGB(163, 190, 140) + Bold,
			SandboxOFF:  FgRGB(76, 86, 106),
			VimNormal:   FgRGB(129, 161, 193) + Bold,
			VimInsert:   FgRGB(163, 190, 140) + Bold,
			VimVisual:   FgRGB(235, 203, 139) + Bold,
			Border:      FgRGB(76, 86, 106),
			Separator:   FgRGB(76, 86, 106),
			Dim:         FgRGB(76, 86, 106),
			Highlight:   FgRGB(255, 255, 255) + Bold,
		}
	case "gruvbox":
		return Palette{
			Name:        "gruvbox",
			Ready:       FgRGB(184, 187, 38) + Bold,
			Thinking:    FgRGB(250, 189, 47) + Bold,
			Working:     FgRGB(142, 192, 124) + Bold,
			Tool:        FgRGB(211, 134, 155) + Bold,
			Init:        FgRGB(131, 165, 152) + Bold,
			Unknown:     FgRGB(235, 219, 178) + Bold,
			Branch:      FgRGB(131, 165, 152) + Bold,
			BranchDirty: FgRGB(251, 73, 52) + Bold,
			DirtyMarker: FgRGB(250, 189, 47) + Bold,
			Model:       FgRGB(211, 134, 155) + Italic,
			ModelMode:   FgRGB(250, 189, 47) + Bold,
			ContextOK:   FgRGB(235, 219, 178),
			ContextWarn: FgRGB(250, 189, 47),
			ContextCrit: FgRGB(251, 73, 52),
			Artifact:    FgRGB(235, 219, 178),
			Subagent:    FgRGB(142, 192, 124) + Bold,
			Task:        FgRGB(235, 219, 178),
			SandboxON:   FgRGB(184, 187, 38) + Bold,
			SandboxOFF:  FgRGB(146, 131, 116),
			VimNormal:   FgRGB(131, 165, 152) + Bold,
			VimInsert:   FgRGB(184, 187, 38) + Bold,
			VimVisual:   FgRGB(250, 189, 47) + Bold,
			Border:      FgRGB(146, 131, 116),
			Separator:   FgRGB(146, 131, 116),
			Dim:         FgRGB(146, 131, 116),
			Highlight:   FgRGB(255, 255, 255) + Bold,
		}
	case "ascii", "plain", "no-color":
		return Palette{
			Name:        "plain",
			Ready:       "",
			Thinking:    "",
			Working:     "",
			Tool:        "",
			Init:        "",
			Unknown:     "",
			Branch:      "",
			BranchDirty: "",
			DirtyMarker: "*",
			Model:       "",
			ModelMode:   "",
			ContextOK:   "",
			ContextWarn: "",
			ContextCrit: "",
			Artifact:    "",
			Subagent:    "",
			Task:        "",
			SandboxON:   "",
			SandboxOFF:  "",
			VimNormal:   "",
			VimInsert:   "",
			VimVisual:   "",
			Border:      "",
			Separator:   "",
			Dim:         "",
			Highlight:   "",
		}
	default:
		return Palette{
			Name:        "default",
			Ready:       FgBrightGreen + Bold,
			Thinking:    FgBrightYellow + Bold,
			Working:     FgBrightCyan + Bold,
			Tool:        FgBrightMagenta + Bold,
			Init:        FgBrightBlue + Bold,
			Unknown:     FgBrightWhite + Bold,
			Branch:      FgBrightBlue + Bold,
			BranchDirty: FgBrightRed + Bold,
			DirtyMarker: FgBrightYellow + Bold,
			Model:       FgBrightMagenta + Italic,
			ModelMode:   FgBrightYellow + Bold,
			ContextOK:   FgBrightWhite,
			ContextWarn: FgBrightYellow,
			ContextCrit: FgBrightRed,
			Artifact:    FgBrightWhite,
			Subagent:    FgBrightCyan + Bold,
			Task:        FgBrightWhite,
			SandboxON:   FgBrightGreen + Bold,
			SandboxOFF:  FgGray,
			VimNormal:   FgBrightBlue + Bold,
			VimInsert:   FgBrightGreen + Bold,
			VimVisual:   FgBrightYellow + Bold,
			Border:      FgGray,
			Separator:   FgGray,
			Dim:         FgGray,
			Highlight:   FgBrightWhite + Bold,
		}
	}
}

type Theme struct {
	Name        string
	ColorActive bool
	Palette     Palette
}

func DefaultTheme() Theme {
	return Theme{
		Name:        "default",
		ColorActive: true,
		Palette:     GetPalette("default"),
	}
}

func PlainTheme() Theme {
	return Theme{
		Name:        "plain",
		ColorActive: false,
		Palette:     GetPalette("plain"),
	}
}

func NamedTheme(name string) Theme {
	p := GetPalette(name)
	isPlain := name == "ascii" || name == "plain" || name == "no-color"
	return Theme{
		Name:        name,
		ColorActive: !isPlain,
		Palette:     p,
	}
}

var AvailableThemes = []string{
	"default",
	"tokyonight",
	"catppuccin",
	"dracula",
	"nord",
	"gruvbox",
	"ascii",
	"plain",
	"no-color",
}
