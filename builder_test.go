package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/widget"
)

func TestBuildBasic(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	root := Build(tree, cfg).Widget()
	if root == nil {
		t.Fatal("expected root widget")
	}
	if root.ElementID() == core.InvalidElementID {
		t.Error("expected valid element ID")
	}
}

func TestBuildWithChildren(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	b := Build(tree, cfg)
	b.Text("hello")
	b.Button("click")

	root := b.Widget()
	if len(root.Children()) != 2 {
		t.Errorf("expected 2 children, got %d", len(root.Children()))
	}
}

func TestBuildRow(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	b := Build(tree, cfg)
	b.Row(8, func(b *Builder) {
		b.Text("a")
		b.Text("b")
	})

	root := b.Widget()
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child (the row), got %d", len(root.Children()))
	}
	row := root.Children()[0]
	if len(row.Children()) != 2 {
		t.Errorf("expected 2 children in row, got %d", len(row.Children()))
	}
}

func TestBuildColumn(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	b := Build(tree, cfg)
	b.Column(12, func(b *Builder) {
		b.Text("x")
		b.Text("y")
		b.Text("z")
	})

	root := b.Widget()
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
	col := root.Children()[0]
	if len(col.Children()) != 3 {
		t.Errorf("expected 3 children in column, got %d", len(col.Children()))
	}
}

func TestBuildNested(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	b := Build(tree, cfg)
	b.Div(func(b *Builder) {
		b.Row(8, func(b *Builder) {
			b.Button("OK").OnClick(func() {})
			b.Button("Cancel")
		})
	})

	root := b.Widget()
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestBuildInput(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	b := Build(tree, cfg)
	b.Input().Placeholder("type here").Value("initial")

	root := b.Widget()
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestBuildCheckbox(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	b := Build(tree, cfg)
	b.Checkbox("agree").Checked(true)

	root := b.Widget()
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}

func TestBuildCustomWidget(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()

	prog := widget.NewProgress(tree, cfg)
	prog.SetPercent(50)

	b := Build(tree, cfg)
	b.Custom(prog)

	root := b.Widget()
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children()))
	}
}
