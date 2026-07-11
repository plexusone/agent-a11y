package remediation

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/plexusone/agent-a11y/types"
	"gopkg.in/yaml.v3"
)

// DesignSystem represents a loaded design system specification.
type DesignSystem struct {
	Name       string                 `yaml:"name"`
	Version    string                 `yaml:"version"`
	Tokens     Tokens                 `yaml:"tokens"`
	Components map[string]interface{} `yaml:"components"`
}

// Tokens contains design token definitions.
type Tokens struct {
	Colors     map[string]ColorToken  `yaml:"colors"`
	Spacing    map[string]string      `yaml:"spacing"`
	Typography map[string]interface{} `yaml:"typography"`
}

// ColorToken represents a color token with metadata.
type ColorToken struct {
	Value       string  `yaml:"value"`
	Description string  `yaml:"description,omitempty"`
	Contrast    float64 `yaml:"contrast,omitempty"` // Pre-calculated contrast vs white
}

// DesignSystemLoader loads and queries design system specs.
type DesignSystemLoader struct {
	ds   *DesignSystem
	path string
}

// LoadDesignSystem loads a design system from a directory.
func LoadDesignSystem(path string) (*DesignSystemLoader, error) {
	// Try common file names
	candidates := []string{
		filepath.Join(path, "design-system.yaml"),
		filepath.Join(path, "design-system.yml"),
		filepath.Join(path, "tokens.yaml"),
		filepath.Join(path, "tokens.yml"),
	}

	var data []byte
	var err error
	var foundPath string

	for _, candidate := range candidates {
		data, err = os.ReadFile(candidate)
		if err == nil {
			foundPath = candidate
			break
		}
	}

	if data == nil {
		return nil, fmt.Errorf("no design system spec found in %s", path)
	}

	var ds DesignSystem
	if err := yaml.Unmarshal(data, &ds); err != nil {
		return nil, fmt.Errorf("failed to parse design system: %w", err)
	}

	return &DesignSystemLoader{
		ds:   &ds,
		path: foundPath,
	}, nil
}

// Info returns design system metadata.
func (l *DesignSystemLoader) Info() *types.DesignSystemInfo {
	return &types.DesignSystemInfo{
		Name:    l.ds.Name,
		Version: l.ds.Version,
		Path:    l.path,
	}
}

// SuggestColorToken suggests a color token that meets contrast requirements.
func (l *DesignSystemLoader) SuggestColorToken(currentValue string, property string, bgColor string, minContrast float64) *types.TokenSuggestion {
	if l.ds.Tokens.Colors == nil {
		return nil
	}

	// Parse current color to RGB
	currentRGB, err := parseColor(currentValue)
	if err != nil {
		return nil
	}

	// Parse background color
	bgRGB, err := parseColor(bgColor)
	if err != nil {
		// Default to white background
		bgRGB = [3]float64{255, 255, 255}
	}

	// Find best matching token
	var bestToken string
	var bestValue string
	var bestContrast float64
	bestDistance := math.MaxFloat64

	for name, token := range l.ds.Tokens.Colors {
		tokenRGB, err := parseColor(token.Value)
		if err != nil {
			continue
		}

		// Calculate contrast against background
		contrast := contrastRatio(tokenRGB, bgRGB)

		// Must meet minimum contrast
		if contrast < minContrast {
			continue
		}

		// Calculate color distance from current (prefer similar colors)
		distance := colorDistance(currentRGB, tokenRGB)

		// Prefer tokens that meet contrast with minimal color change
		if distance < bestDistance || (distance == bestDistance && contrast > bestContrast) {
			bestToken = name
			bestValue = token.Value
			bestContrast = contrast
			bestDistance = distance
		}
	}

	if bestToken == "" {
		return nil
	}

	return &types.TokenSuggestion{
		Property:      property,
		CurrentValue:  currentValue,
		TokenName:     bestToken,
		TokenValue:    bestValue,
		Rationale:     fmt.Sprintf("Meets %.1f:1 contrast ratio (min: %.1f:1)", bestContrast, minContrast),
		ContrastRatio: bestContrast,
	}
}

// SuggestSpacingToken suggests a spacing token.
func (l *DesignSystemLoader) SuggestSpacingToken(currentValue string, property string) *types.TokenSuggestion {
	if l.ds.Tokens.Spacing == nil {
		return nil
	}

	// Parse current value to pixels
	currentPx := parsePixelValue(currentValue)
	if currentPx < 0 {
		return nil
	}

	// Find closest token
	var bestToken string
	var bestValue string
	bestDiff := math.MaxFloat64

	for name, value := range l.ds.Tokens.Spacing {
		tokenPx := parsePixelValue(value)
		if tokenPx < 0 {
			continue
		}

		diff := math.Abs(currentPx - tokenPx)
		if diff < bestDiff {
			bestToken = name
			bestValue = value
			bestDiff = diff
		}
	}

	if bestToken == "" {
		return nil
	}

	return &types.TokenSuggestion{
		Property:     property,
		CurrentValue: currentValue,
		TokenName:    bestToken,
		TokenValue:   bestValue,
		Rationale:    "Use design system spacing token",
	}
}

// DetectComponent attempts to identify a design system component from HTML.
func (l *DesignSystemLoader) DetectComponent(html string, tagName string, classes []string) (componentID string, variant string) {
	if l.ds.Components == nil {
		return "", ""
	}

	// Simple heuristic: check if tag name or class matches component names
	lowerTag := strings.ToLower(tagName)

	for compID := range l.ds.Components {
		lowerComp := strings.ToLower(compID)

		// Direct tag match
		if lowerTag == lowerComp {
			return compID, ""
		}

		// Class contains component name
		for _, class := range classes {
			lowerClass := strings.ToLower(class)
			if strings.Contains(lowerClass, lowerComp) {
				// Try to extract variant from class
				variant := extractVariant(lowerClass, lowerComp)
				return compID, variant
			}
		}
	}

	return "", ""
}

// Helper functions

func parseColor(color string) ([3]float64, error) {
	color = strings.TrimSpace(color)

	// Handle hex colors
	if strings.HasPrefix(color, "#") {
		hex := strings.TrimPrefix(color, "#")
		if len(hex) == 3 {
			// Expand shorthand
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		if len(hex) != 6 {
			return [3]float64{}, fmt.Errorf("invalid hex color: %s", color)
		}

		var rgb [3]float64
		for i := 0; i < 3; i++ {
			var val int
			_, err := fmt.Sscanf(hex[i*2:i*2+2], "%02x", &val)
			if err != nil {
				return [3]float64{}, err
			}
			rgb[i] = float64(val)
		}
		return rgb, nil
	}

	// Handle rgb()/rgba()
	if strings.HasPrefix(color, "rgb") {
		// Strip rgb()/rgba() wrapper
		inner := strings.TrimPrefix(color, "rgba(")
		inner = strings.TrimPrefix(inner, "rgb(")
		inner = strings.TrimSuffix(inner, ")")

		parts := strings.Split(inner, ",")
		if len(parts) < 3 {
			return [3]float64{}, fmt.Errorf("invalid rgb color: %s", color)
		}

		var rgb [3]float64
		for i := 0; i < 3; i++ {
			var val float64
			_, err := fmt.Sscanf(strings.TrimSpace(parts[i]), "%f", &val)
			if err != nil {
				return [3]float64{}, err
			}
			rgb[i] = val
		}
		return rgb, nil
	}

	return [3]float64{}, fmt.Errorf("unsupported color format: %s", color)
}

func luminance(rgb [3]float64) float64 {
	// sRGB to linear conversion and luminance calculation
	var linear [3]float64
	for i, val := range rgb {
		v := val / 255.0
		if v <= 0.03928 {
			linear[i] = v / 12.92
		} else {
			linear[i] = math.Pow((v+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*linear[0] + 0.7152*linear[1] + 0.0722*linear[2]
}

func contrastRatio(fg, bg [3]float64) float64 {
	l1 := luminance(fg)
	l2 := luminance(bg)

	lighter := math.Max(l1, l2)
	darker := math.Min(l1, l2)

	return (lighter + 0.05) / (darker + 0.05)
}

func colorDistance(c1, c2 [3]float64) float64 {
	// Simple Euclidean distance in RGB space
	dr := c1[0] - c2[0]
	dg := c1[1] - c2[1]
	db := c1[2] - c2[2]
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func parsePixelValue(value string) float64 {
	value = strings.TrimSpace(value)

	// Handle px values
	if strings.HasSuffix(value, "px") {
		var px float64
		_, err := fmt.Sscanf(value, "%fpx", &px)
		if err != nil {
			return -1
		}
		return px
	}

	// Handle rem values (assume 16px base)
	if strings.HasSuffix(value, "rem") {
		var rem float64
		_, err := fmt.Sscanf(value, "%frem", &rem)
		if err != nil {
			return -1
		}
		return rem * 16
	}

	// Handle em values (assume 16px base)
	if strings.HasSuffix(value, "em") {
		var em float64
		_, err := fmt.Sscanf(value, "%fem", &em)
		if err != nil {
			return -1
		}
		return em * 16
	}

	// Handle plain numbers (assume px)
	var num float64
	_, err := fmt.Sscanf(value, "%f", &num)
	if err != nil {
		return -1
	}
	return num
}

func extractVariant(class, component string) string {
	// Common variant patterns: button-primary, btn--secondary, Button_variant
	class = strings.ToLower(class)
	component = strings.ToLower(component)

	// Remove component name
	remaining := strings.TrimPrefix(class, component)
	remaining = strings.TrimPrefix(remaining, "-")
	remaining = strings.TrimPrefix(remaining, "--")
	remaining = strings.TrimPrefix(remaining, "_")

	// Take first word as variant
	parts := strings.FieldsFunc(remaining, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})

	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
