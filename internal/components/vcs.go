package components

import (
	"github.com/joaco/antigravity-statusline/internal/config"
	"github.com/joaco/antigravity-statusline/internal/payload"
	"github.com/joaco/antigravity-statusline/internal/theme"
)

func RenderVCSBadge(vcs payload.VCSInfo, p theme.Palette, icons config.IconsConfig) string {
	if vcs.Branch == "" {
		return ""
	}

	branch := payload.Sanitize(vcs.Branch)
	if branch == "" {
		return ""
	}

	icon := theme.IconGitBranch
	if icons.Branch != "" {
		icon = icons.Branch
	}

	dirtyIcon := theme.IconGitDirty
	if icons.Dirty != "" {
		dirtyIcon = icons.Dirty
	}

	branchColor := p.Branch
	if vcs.Dirty {
		branchColor = p.BranchDirty
	}

	branchText := theme.Style(branchColor, icon+" "+branch)

	if !vcs.Dirty {
		return branchText
	}

	dirtyMarker := theme.Style(p.DirtyMarker, dirtyIcon)
	return branchText + dirtyMarker
}
