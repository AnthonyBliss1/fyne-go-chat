package ui

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
)

//go:embed Tabular/Fonts/OTF/Tabular-Regular.otf
var tabularBytes []byte

var ResourceTabular = fyne.NewStaticResource("Tabular-Regular.otf", tabularBytes)

type ForcedVariant struct {
	fyne.Theme
	Variant fyne.ThemeVariant
}

func (f *ForcedVariant) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.Variant)
}

func (f *ForcedVariant) Font(s fyne.TextStyle) fyne.Resource {
	return ResourceTabular
}

func (f *ForcedVariant) Icon(name fyne.ThemeIconName) fyne.Resource {
	return f.Theme.Icon(name)
}
func (f *ForcedVariant) Size(name fyne.ThemeSizeName) float32 {
	return f.Theme.Size(name)
}
