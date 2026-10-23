package gallery

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/theme"
)

func TestButtonPreview(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, width := range []int{540, 280} {
			g := &Gallery{}
			example, _ := controlExampleFor(pageFromKey("controls/Button"))
			tt := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(theme.Palette(dark))
				g.Controls(c, example)
			}, width, 600)
			if _, ok := tt.Find("Button"); !ok || tt.HasText("Default") {
				t.Fatal("button section must use the component name")
			}
			for _, label := range []string{"Primary", "Secondary", "Outline", "Ghost", "Link"} {
				r, ok := tt.Find(label)
				if !ok || r.X < 20 || r.X+r.W > float32(width)-20 {
					t.Fatalf("%s overflows preview: %+v", label, r)
				}
				if err := tt.Click(label); err != nil {
					t.Fatal(err)
				}
			}
			if g.controlState("Button", "Button").button.clicks != [5]int{1, 1, 1, 1, 1} {
				t.Fatal("preview lost button interactions")
			}
		}
	}
}

func TestButtonTogglePreview(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, width := range []int{540, 280} {
			g := &Gallery{}
			tt := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(theme.Palette(dark))
				example, _ := controlExampleFor(pageFromKey("controls/Button"))
				g.Controls(c, example)
			}, width, 1000)
			if !tt.HasText("ToggleButton") || tt.HasText("Disabled") {
				t.Fatal("missing toggle section or unexpected disabled preview")
			}
			s := g.controlState("Button", "ToggleButton")
			if s.on || !s.toggles[1] {
				t.Fatal("preview must demonstrate off and on states")
			}
			for _, label := range []string{"Bold", "Italic"} {
				r, ok := tt.Find(label)
				if !ok || r.X < 20 || r.X+r.W > float32(width)-20 {
					t.Fatalf("%s overflows preview: %+v", label, r)
				}
				if err := tt.Click(label); err != nil {
					t.Fatal(err)
				}
			}
			if !s.on || s.toggles[1] {
				t.Fatal("toggle preview did not update independent values")
			}
			if g.controlState("Button", "Button").button.clicks != [5]int{} {
				t.Fatal("toggle state leaked into other button examples")
			}
			if pageFromKey("controls/ToggleButton") != lucidePage {
				t.Fatal("standalone toggle route remains")
			}
		}
	}
}

func TestControlButtonWrap(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, width := range []int{540, 200} {
			s := &controlState{}
			tt := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(theme.Palette(dark))
				ui.Column(c).Fill().Gap(12).Children(func() { controlButton(c, s, "Default") })
			}, width, 300)
			var previous ui.Rect
			for i, label := range []string{"Primary", "Secondary", "Outline", "Ghost", "Link"} {
				r, ok := tt.Find(label)
				if !ok {
					t.Fatalf("missing button %s", label)
				}
				if r.X < 0 || r.X+r.W > float32(width)+1 {
					t.Fatalf("%s overflows width %d: %+v", label, width, r)
				}
				if i > 0 {
					if width == 540 && r.Y+r.H/2 != previous.Y+previous.H/2 {
						t.Fatalf("%s is not on the same row", label)
					}
					if r.Y < previous.Y+previous.H && r.X < previous.X+previous.W {
						t.Fatalf("%s overlaps the previous button", label)
					}
				}
				previous = r
				if err := tt.Click(label); err != nil {
					t.Fatal(err)
				}
			}
			if s.button.clicks != [5]int{1, 1, 1, 1, 1} {
				t.Fatalf("unexpected per-button clicks: %v", s.button.clicks)
			}
		}
	}
}

func TestControlButtonPointerStyles(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, disabled := range []bool{false, true} {
			s := &controlState{}
			tt := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(theme.Palette(dark))
				ui.Column(c).Fill().Disabled(disabled).Children(func() { controlButton(c, s, "Default") })
			}, 540, 300)
			for _, label := range []string{"Primary", "Secondary", "Outline", "Ghost"} {
				r, ok := tt.Find(label)
				if !ok {
					t.Fatalf("missing button %s", label)
				}
				x, y := r.X+6, r.Y+r.H/2
				tt.Move(530, 290)
				rest := tt.Image().RGBAAt(int(x), int(y))
				tt.Move(x, y)
				hover := tt.Image().RGBAAt(int(x), int(y))
				tt.Press(x, y)
				pressed := tt.Image().RGBAAt(int(x), int(y))
				tt.Release(x, y)
				tt.Move(530, 290)
				reset := tt.Image().RGBAAt(int(x), int(y))
				if disabled {
					if hover != rest || pressed != rest {
						t.Fatalf("%v/%s: disabled button changed on pointer input", dark, label)
					}
				} else if label == "Primary" {
					// The primary keeps its accent fill on hover: its lift
					// is a shadow the pixel sampling cannot see. The press
					// still darkens the fill.
					if pressed == hover {
						t.Fatalf("%v/%s: missing pressed feedback", dark, label)
					}
				} else if hover == rest || pressed == hover {
					t.Fatalf("%v/%s: missing hover or pressed feedback", dark, label)
				}
				if reset != rest {
					t.Fatalf("%v/%s: pointer exit did not restore resting style", dark, label)
				}
			}
		}
	}
}

func TestButtonIndependentCounts(t *testing.T) {
	s := &controlState{}
	tt := ui.NewTester(func(c *ui.Context) {
		controlButton(c, s, "Button")
	}, 540, 300)
	labels := []string{"Primary", "Secondary", "Outline", "Ghost", "Link"}
	for i, label := range labels {
		for range i + 1 {
			if err := tt.Click(label); err != nil {
				t.Fatal(err)
			}
		}
		for j, name := range labels {
			want := 0
			if j <= i {
				want = j + 1
			}
			if s.button.clicks[j] != want {
				t.Fatalf("%s count = %d, want %d", name, s.button.clicks[j], want)
			}
			// MyGo exposes Link's inline text through its accessible label,
			// not as a separate text rectangle.
			if name == "Link" {
				continue
			}
			content, ok := tt.Find(fmt.Sprintf("%s: %d", name, want))
			button, found := tt.Find(name)
			if !ok || !found || content.X < button.X || content.Y < button.Y || content.X+content.W > button.X+button.W || content.Y+content.H > button.Y+button.H {
				t.Fatalf("%s count is not inside its button: content=%+v (%t), button=%+v (%t)", name, content, ok, button, found)
			}
		}
	}
	if tt.HasText("Clicks:") {
		t.Fatal("unexpected shared click counter")
	}
}

func TestButtonIconAndToolbarPreview(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, width := range []int{540, 280} {
			g := &Gallery{}
			example, _ := controlExampleFor(pageFromKey("controls/Button"))
			tt := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(theme.Palette(dark))
				g.Controls(c, example)
			}, width, 1400)
			link, _ := tt.Find("Link")
			toolbar, ok := tt.Find("Filter")
			if !ok || toolbar.Y < link.Y+link.H {
				t.Fatalf("toolbar button must sit below the standard buttons: %+v, %+v", toolbar, link)
			}
			for _, section := range []string{"Icon and text", "IconButton", "Shell toolbar"} {
				if !tt.HasText(section) {
					t.Fatalf("missing %s example", section)
				}
			}
			for _, label := range []string{"New", "Edit", "Copy", "Favorite", "Filter", "Upload", "Redo", "New folder", "Settings"} {
				r, ok := tt.Find(label)
				if !ok || r.X < 20 || r.X+r.W > float32(width)-20 {
					t.Fatalf("%s overflows preview: %+v", label, r)
				}
				if err := tt.Click(label); err != nil {
					t.Fatal(err)
				}
			}
			actions := g.controlState("Button", "Toolbar").button.actions
			if actions["filter"] != 1 || actions["upload"] != 1 || actions["redo"] != 1 || actions["newFolder"] != 1 || actions["settings"] != 1 {
				t.Fatalf("toolbar actions = %v", actions)
			}
			_ = tt.Click("Delete")
			if g.controlState("Button", "IconButton").button.actions["delete"] != 0 {
				t.Fatal("disabled icon button reported a click")
			}
		}
	}
}
