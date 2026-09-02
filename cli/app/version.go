package app

import (
	"github.com/yumauri/fbrcm/core/about"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
)

const versionLine = `{{with .Name}}{{printf "%s " .}}{{end}}{{printf "%s\n" .Version}}`

func buildVersionTemplate() string {
	plain := about.Logo + "\n"
	logo := plain
	if !clistyles.NoColorEnabled() {
		// Machine-readable version output carries text, but never terminal styling.
		logo = `{{if (.Flags.GetBool "json")}}` + plain + "{{else}}" + about.RenderLogo(true) + "\n{{end}}"
	}
	return logo + versionLine + about.Author + "\n"
}
