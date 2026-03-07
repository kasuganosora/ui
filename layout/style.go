package layout

// Display determines the layout mode of an element.
type Display uint8

const (
	DisplayBlock   Display = iota // Default (zero value)
	DisplayFlex                   // Flexbox container
	DisplayInline                 // Inline flow
	DisplayNone                   // Hidden, removed from layout
)

// Position determines positioning behavior.
type Position uint8

const (
	PositionRelative Position = iota
	PositionAbsolute
)

// FlexDirection determines the main axis direction.
type FlexDirection uint8

const (
	FlexDirectionRow FlexDirection = iota
	FlexDirectionColumn
	FlexDirectionRowReverse
	FlexDirectionColumnReverse
)

// FlexWrap determines whether flex items wrap.
type FlexWrap uint8

const (
	FlexWrapNoWrap FlexWrap = iota
	FlexWrapWrap
	FlexWrapWrapReverse
)

// JustifyContent determines main-axis alignment.
type JustifyContent uint8

const (
	JustifyFlexStart    JustifyContent = iota
	JustifyFlexEnd
	JustifyCenter
	JustifySpaceBetween
	JustifySpaceAround
	JustifySpaceEvenly
)

// AlignItems determines cross-axis alignment of items.
type AlignItems uint8

const (
	AlignStretch  AlignItems = iota
	AlignFlexStart
	AlignFlexEnd
	AlignCenter
	AlignBaseline
)

// AlignSelf overrides the parent's AlignItems for a single item.
type AlignSelf uint8

const (
	AlignSelfAuto AlignSelf = iota
	AlignSelfStretch
	AlignSelfFlexStart
	AlignSelfFlexEnd
	AlignSelfCenter
	AlignSelfBaseline
)

// AlignContent determines cross-axis alignment of flex lines.
type AlignContent uint8

const (
	AlignContentStretch    AlignContent = iota
	AlignContentFlexStart
	AlignContentFlexEnd
	AlignContentCenter
	AlignContentSpaceBetween
	AlignContentSpaceAround
)

// Overflow determines how overflow content is handled.
type Overflow uint8

const (
	OverflowVisible Overflow = iota
	OverflowHidden
	OverflowScroll
)

// Value represents a CSS dimension value. Value object.
// Zero value means "auto" (Undefined).
type Value struct {
	Amount float32
	Unit   Unit
}

// Unit is the type of a dimension value.
type Unit uint8

const (
	UnitUndefined Unit = iota // auto
	UnitPx                    // pixels
	UnitPercent               // percentage of parent
)

// Predefined values.
var (
	Auto    = Value{Unit: UnitUndefined}
	Zero    = Value{Amount: 0, Unit: UnitPx}
)

// Px creates a pixel value.
func Px(v float32) Value { return Value{Amount: v, Unit: UnitPx} }

// Pct creates a percentage value.
func Pct(v float32) Value { return Value{Amount: v, Unit: UnitPercent} }

// IsAuto returns true if the value is undefined/auto.
func (v Value) IsAuto() bool { return v.Unit == UnitUndefined }

// Resolve resolves the value against a parent dimension.
// Returns (resolved, isDefined).
func (v Value) Resolve(parentSize float32) (float32, bool) {
	switch v.Unit {
	case UnitPx:
		return v.Amount, true
	case UnitPercent:
		return v.Amount / 100 * parentSize, true
	default:
		return 0, false
	}
}

// EdgeValues represents four-sided values (margin, padding, border).
type EdgeValues struct {
	Top, Right, Bottom, Left Value
}

// Style contains all layout-relevant CSS properties for an element.
// Value object — created by the style resolver for each element.
type Style struct {
	// Box model
	Display  Display
	Position Position
	Overflow Overflow

	// Dimensions
	Width    Value
	Height   Value
	MinWidth  Value
	MinHeight Value
	MaxWidth  Value
	MaxHeight Value

	// Spacing
	Margin  EdgeValues
	Padding EdgeValues
	Border  EdgeValues

	// Positioning (for Position: Absolute)
	Top    Value
	Right  Value
	Bottom Value
	Left   Value

	// Flexbox container properties
	FlexDirection  FlexDirection
	FlexWrap       FlexWrap
	JustifyContent JustifyContent
	AlignItems     AlignItems
	AlignContent   AlignContent
	Gap            float32 // Gap between flex items (px)
	RowGap         float32 // Row gap for wrapped flex (px), 0 = use Gap
	ColumnGap      float32 // Column gap (px), 0 = use Gap

	// Flexbox item properties
	FlexGrow   float32
	FlexShrink float32
	FlexBasis  Value
	AlignSelf  AlignSelf
	Order      int
}

// DefaultStyle returns the default style (block display, auto sizing).
func DefaultStyle() Style {
	return Style{
		Display:    DisplayBlock,
		FlexShrink: 1, // CSS default
		AlignItems: AlignStretch,
	}
}

// MainGap returns the gap for the main axis.
func (s *Style) MainGap() float32 {
	if s.FlexDirection == FlexDirectionRow || s.FlexDirection == FlexDirectionRowReverse {
		if s.ColumnGap > 0 {
			return s.ColumnGap
		}
		return s.Gap
	}
	if s.RowGap > 0 {
		return s.RowGap
	}
	return s.Gap
}

// CrossGap returns the gap for the cross axis.
func (s *Style) CrossGap() float32 {
	if s.FlexDirection == FlexDirectionRow || s.FlexDirection == FlexDirectionRowReverse {
		if s.RowGap > 0 {
			return s.RowGap
		}
		return s.Gap
	}
	if s.ColumnGap > 0 {
		return s.ColumnGap
	}
	return s.Gap
}

// IsRow returns true if the main axis is horizontal.
func (s *Style) IsRow() bool {
	return s.FlexDirection == FlexDirectionRow || s.FlexDirection == FlexDirectionRowReverse
}

// IsReverse returns true if the direction is reversed.
func (s *Style) IsReverse() bool {
	return s.FlexDirection == FlexDirectionRowReverse || s.FlexDirection == FlexDirectionColumnReverse
}
