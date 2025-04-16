package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type CustomTheme struct {
    fyne.Theme
}

func NewCustomTheme() fyne.Theme {
    return &CustomTheme{Theme: theme.DefaultTheme()}
}

func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
    switch name {
    case theme.ColorNameBackground:
        return color.NRGBA{R: 240, G: 240, B: 250, A: 255} // Light blue-gray
    case theme.ColorNameForeground:
        return color.NRGBA{R: 33, G: 33, B: 33, A: 255} // Dark gray for text
    case theme.ColorNameInputBackground:
        return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // Pure white for input fields
    case theme.ColorNameButton:
        return color.NRGBA{R: 43, G: 90, B: 151, A: 255} // #2b5a97 - Matching logo color
    case theme.ColorNameShadow:
        return color.NRGBA{R: 0, G: 0, B: 0, A: 40} // Subtle shadow
    case theme.ColorNamePlaceHolder:
        return color.NRGBA{R: 128, G: 128, B: 128, A: 255} // Medium gray for placeholders
    case theme.ColorNamePrimary:
        return color.NRGBA{R: 43, G: 90, B: 151, A: 255} // Primary color matching button
    case theme.ColorNameHover:
        return color.NRGBA{R: 33, G: 75, B: 126, A: 255} // Slightly darker for hover states
    case theme.ColorNameInputBorder:
        return color.NRGBA{R: 200, G: 200, B: 200, A: 255} // Light gray for input borders
    }
    return t.Theme.Color(name, variant)
}

// Optional: Customize text size if needed
func (t *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
    switch name {
    case theme.SizeNameText:
        return 14
    case theme.SizeNameHeadingText:
        return 20
    }
    return t.Theme.Size(name)
}
