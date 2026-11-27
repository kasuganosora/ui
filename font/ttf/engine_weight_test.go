package ttf

import (
	"testing"

	"github.com/kasuganosora/ui/font"
	"golang.org/x/image/font/gofont/goregular"
)

func TestWeightTokenMapping(t *testing.T) {
	cases := []struct {
		weight font.Weight
		want   string
	}{
		{font.WeightThin, "thin"},
		{font.WeightLight, "light"},
		{font.WeightRegular, "regular"},
		{font.WeightMedium, "medium"},
		{font.WeightSemiBold, "semibold"},
		{font.WeightBold, "bold"},
		{font.Weight(451), ""}, // arbitrary variable-font value: do not guess a face
	}
	for _, tc := range cases {
		if got := weightToken(tc.weight); got != tc.want {
			t.Errorf("weightToken(%d) = %q, want %q", tc.weight, got, tc.want)
		}
	}
}

// Whole-word matching matters: a request for Bold must not silently pick a
// Semibold face, and localized subfamily names carry no English weight word.
func TestWeightMatchesWholeWords(t *testing.T) {
	cases := []struct {
		subfamily string
		token     string
		want      bool
	}{
		{"Regular", "regular", true},
		{"Medium", "medium", true},
		{"Semibold", "semibold", true},
		{"Semibold", "bold", false},
		{"Ultralight", "light", false},
		{"中黑體", "medium", false},
		{"", "medium", false},
	}
	for _, tc := range cases {
		if got := weightMatches(tc.subfamily, tc.token); got != tc.want {
			t.Errorf("weightMatches(%q,%q) = %v, want %v", tc.subfamily, tc.token, got, tc.want)
		}
	}
}

// A plain single-font file has only one face, so a weighted load must still
// succeed (it is a visual no-op there), and malformed data must still error
// instead of being accepted as a new face.
func TestWeightedLoadOnSingleFontAndInvalidData(t *testing.T) {
	e := New()
	defer e.Destroy()

	if _, err := e.LoadFontWithWeight([]byte("not a font"), font.WeightMedium); err == nil {
		t.Error("weighted load accepted malformed font data")
	}

	plain, err := e.LoadFont(goregular.TTF)
	if err != nil {
		t.Fatalf("plain load: %v", err)
	}
	weighted, err := e.LoadFontWithWeight(goregular.TTF, font.WeightMedium)
	if err != nil {
		t.Fatalf("weighted load of a single font should degrade, not fail: %v", err)
	}
	if plain == weighted {
		t.Fatal("weighted load must return its own face ID")
	}
	inkOf := func(id font.ID) float64 {
		bmp, err := e.RasterizeGlyph(id, e.GlyphIndex(id, 'A'), 12, false)
		if err != nil {
			t.Fatalf("rasterize: %v", err)
		}
		sum := 0.0
		for _, v := range bmp.Data {
			sum += float64(v)
		}
		return sum
	}
	if inkOf(plain) != inkOf(weighted) {
		t.Error("single-font weighted load changed the glyph outline")
	}
}

// compile-time assertion: the pure Go engine implements the optional
// weight-selection capability that the application shell consumes.
var _ font.WeightedLoader = (*Engine)(nil)
