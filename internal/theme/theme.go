package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a complete color theme for the application
type Theme struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Primary     lipgloss.AdaptiveColor `json:"primary"`
	Success     lipgloss.AdaptiveColor `json:"success"`
	Danger      lipgloss.AdaptiveColor `json:"danger"`
	Warning     lipgloss.AdaptiveColor `json:"warning"`
	Muted       lipgloss.AdaptiveColor `json:"muted"`
	Background  lipgloss.AdaptiveColor `json:"background"`
	Text        lipgloss.AdaptiveColor `json:"text"`
	Border      lipgloss.AdaptiveColor `json:"border"`
}

// Predefined themes following Charm design patterns
var (
	// DefaultTheme - Current blue theme (maintains existing look)
	DefaultTheme = Theme{
		ID:          "default",
		Name:        "Default",
		Description: "Classic blue theme with professional styling",
		Primary:     lipgloss.AdaptiveColor{Light: "#0EA5E9", Dark: "#0EA5E9"},
		Success:     lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#10B981"},
		Danger:      lipgloss.AdaptiveColor{Light: "#EF4444", Dark: "#EF4444"},
		Warning:     lipgloss.AdaptiveColor{Light: "#F59E0B", Dark: "#F59E0B"},
		Muted:       lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#6B7280"},
		Background:  lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#111827"},
		Text:        lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#F9FAFB"},
		Border:      lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#334155"},
	}

	// MonochromeTheme - Professional grayscale theme
	MonochromeTheme = Theme{
		ID:          "monochrome",
		Name:        "Monochrome",
		Description: "Elegant grayscale theme for distraction-free work",
		Primary:     lipgloss.AdaptiveColor{Light: "#374151", Dark: "#9CA3AF"},
		Success:     lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#D1D5DB"},
		Danger:      lipgloss.AdaptiveColor{Light: "#4B5563", Dark: "#F3F4F6"},
		Warning:     lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#E5E7EB"},
		Muted:       lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"},
		Background:  lipgloss.AdaptiveColor{Light: "#F9FAFB", Dark: "#1F2937"},
		Text:        lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F9FAFB"},
		Border:      lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"},
	}

	// SolarizedTheme - Warm, eye-friendly Solarized color scheme
	SolarizedTheme = Theme{
		ID:          "solarized",
		Name:        "Solarized",
		Description: "Warm, scientifically-designed color palette",
		Primary:     lipgloss.AdaptiveColor{Light: "#268BD2", Dark: "#268BD2"},
		Success:     lipgloss.AdaptiveColor{Light: "#859900", Dark: "#859900"},
		Danger:      lipgloss.AdaptiveColor{Light: "#DC322F", Dark: "#DC322F"},
		Warning:     lipgloss.AdaptiveColor{Light: "#B58900", Dark: "#B58900"},
		Muted:       lipgloss.AdaptiveColor{Light: "#93A1A1", Dark: "#586E75"},
		Background:  lipgloss.AdaptiveColor{Light: "#FDF6E3", Dark: "#002B36"},
		Text:        lipgloss.AdaptiveColor{Light: "#657B83", Dark: "#839496"},
		Border:      lipgloss.AdaptiveColor{Light: "#EEE8D5", Dark: "#073642"},
	}

	// DraculaTheme - Popular dark theme with purple accents
	DraculaTheme = Theme{
		ID:          "dracula",
		Name:        "Dracula",
		Description: "Dark theme with vibrant purple and pink accents",
		Primary:     lipgloss.AdaptiveColor{Light: "#6272A4", Dark: "#BD93F9"},
		Success:     lipgloss.AdaptiveColor{Light: "#50FA7B", Dark: "#50FA7B"},
		Danger:      lipgloss.AdaptiveColor{Light: "#FF5555", Dark: "#FF5555"},
		Warning:     lipgloss.AdaptiveColor{Light: "#FFB86C", Dark: "#FFB86C"},
		Muted:       lipgloss.AdaptiveColor{Light: "#6272A4", Dark: "#6272A4"},
		Background:  lipgloss.AdaptiveColor{Light: "#F8F8F2", Dark: "#282A36"},
		Text:        lipgloss.AdaptiveColor{Light: "#44475A", Dark: "#F8F8F2"},
		Border:      lipgloss.AdaptiveColor{Light: "#6272A4", Dark: "#44475A"},
	}

	// NordTheme - Cool, arctic-inspired color palette
	NordTheme = Theme{
		ID:          "nord",
		Name:        "Nord", 
		Description: "Arctic-inspired color palette with cool blues",
		Primary:     lipgloss.AdaptiveColor{Light: "#5E81AC", Dark: "#88C0D0"},
		Success:     lipgloss.AdaptiveColor{Light: "#A3BE8C", Dark: "#A3BE8C"},
		Danger:      lipgloss.AdaptiveColor{Light: "#BF616A", Dark: "#BF616A"},
		Warning:     lipgloss.AdaptiveColor{Light: "#EBCB8B", Dark: "#EBCB8B"},
		Muted:       lipgloss.AdaptiveColor{Light: "#4C566A", Dark: "#4C566A"},
		Background:  lipgloss.AdaptiveColor{Light: "#ECEFF4", Dark: "#2E3440"},
		Text:        lipgloss.AdaptiveColor{Light: "#2E3440", Dark: "#ECEFF4"},
		Border:      lipgloss.AdaptiveColor{Light: "#D8DEE9", Dark: "#3B4252"},
	}

	// GruvboxMaterialTheme - Warm, earthy colors designed to be easy on the eyes
	GruvboxMaterialTheme = Theme{
		ID:          "gruvbox-material",
		Name:        "Gruvbox Material",
		Description: "Warm, earthy theme designed to protect developers' eyes",
		Primary:     lipgloss.AdaptiveColor{Light: "#d4be98", Dark: "#ebdbb2"}, // Light beige / Cream (swapped from border)
		Success:     lipgloss.AdaptiveColor{Light: "#a9b665", Dark: "#a9b665"}, // Green
		Danger:      lipgloss.AdaptiveColor{Light: "#ea6962", Dark: "#ea6962"}, // Red
		Warning:     lipgloss.AdaptiveColor{Light: "#d8a657", Dark: "#d8a657"}, // Yellow
		Muted:       lipgloss.AdaptiveColor{Light: "#7daea3", Dark: "#7daea3"}, // Blue-green (swapped from primary)
		Background:  lipgloss.AdaptiveColor{Light: "#fbf1c7", Dark: "#282828"}, // Light cream / Dark brown
		Text:        lipgloss.AdaptiveColor{Light: "#3c3836", Dark: "#d4be98"}, // Dark brown / Light beige
		Border:      lipgloss.AdaptiveColor{Light: "#928374", Dark: "#504945"}, // Gray / Dark gray (swapped from muted)
	}
)

// GetAllThemes returns all available themes
func GetAllThemes() []Theme {
	return []Theme{
		DefaultTheme,
		MonochromeTheme,
		SolarizedTheme,
		DraculaTheme,
		NordTheme,
		GruvboxMaterialTheme,
	}
}

// GetThemeByID returns a theme by its ID, defaults to DefaultTheme if not found
func GetThemeByID(id string) Theme {
	for _, theme := range GetAllThemes() {
		if theme.ID == id {
			return theme
		}
	}
	return DefaultTheme
}

// GetThemeNames returns a slice of all theme names for UI display
func GetThemeNames() []string {
	themes := GetAllThemes()
	names := make([]string, len(themes))
	for i, theme := range themes {
		names[i] = theme.Name
	}
	return names
}

// ThemePreview represents a preview of theme colors for UI display
type ThemePreview struct {
	Theme    Theme
	ColorBar string // Rendered color bar showing theme colors
}

// GeneratePreview creates a visual preview of the theme colors
func (t Theme) GeneratePreview() ThemePreview {
	// Create a color bar showing primary colors
	primaryBlock := lipgloss.NewStyle().Background(t.Primary).Render("   ")
	successBlock := lipgloss.NewStyle().Background(t.Success).Render("   ")
	dangerBlock := lipgloss.NewStyle().Background(t.Danger).Render("   ")
	warningBlock := lipgloss.NewStyle().Background(t.Warning).Render("   ")
	
	colorBar := primaryBlock + successBlock + dangerBlock + warningBlock
	
	return ThemePreview{
		Theme:    t,
		ColorBar: colorBar,
	}
}