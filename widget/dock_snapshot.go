package widget

import (
	"encoding/json"
	"fmt"
)

// DockPanelState is one node in a persisted dock tree.
type DockPanelState struct {
	ID       string
	Title    string
	Position DockPosition
	DockSize float32
	FloatX   float32
	FloatY   float32
	FloatW   float32
	FloatH   float32
	Visible  bool
}

// DockSnapshot is the layout tree: panel identity, geometry, tab order, and scale.
type DockSnapshot struct {
	Panels   []DockPanelState
	Center   []string
	ActiveID string
	Scale    float32
}

// Snapshot copies the current layout. The caller can store the JSON form.
func (dl *DockLayout) Snapshot() DockSnapshot {
	s := DockSnapshot{
		Panels:   make([]DockPanelState, 0, len(dl.panels)),
		Center:   append([]string(nil), dl.centerTabs...),
		ActiveID: dl.ActiveID(),
		Scale:    dl.scaleOr1(),
	}
	for _, p := range dl.panels {
		if p == nil {
			continue
		}
		s.Panels = append(s.Panels, DockPanelState{
			ID: p.ID, Title: p.Title, Position: p.Position, DockSize: p.DockSize,
			FloatX: p.FloatX, FloatY: p.FloatY, FloatW: p.FloatW, FloatH: p.FloatH,
			Visible: p.Visible,
		})
	}
	return s
}

// MarshalLayout encodes Snapshot as JSON.
func (dl *DockLayout) MarshalLayout() ([]byte, error) {
	return json.Marshal(dl.Snapshot())
}

// UnmarshalLayout validates b and replaces the layout.
// An invalid payload leaves the current panels, scale, and active tab in place.
func (dl *DockLayout) UnmarshalLayout(b []byte) error {
	var s DockSnapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return dl.Restore(s)
}

// Restore applies a validated snapshot. Content widgets are kept when the ID matches.
func (dl *DockLayout) Restore(s DockSnapshot) error {
	if err := validateDock(s); err != nil {
		return err
	}
	content := map[string]Widget{}
	for _, p := range dl.panels {
		if p != nil && p.Content != nil {
			content[p.ID] = p.Content
		}
	}
	panels := make([]*DockPanel, len(s.Panels))
	for i, st := range s.Panels {
		panels[i] = &DockPanel{
			ID: st.ID, Title: st.Title, Content: content[st.ID], Position: st.Position,
			DockSize: st.DockSize, FloatX: st.FloatX, FloatY: st.FloatY,
			FloatW: st.FloatW, FloatH: st.FloatH, Visible: st.Visible,
		}
	}
	dl.panels = panels
	dl.centerTabs = append([]string(nil), s.Center...)
	dl.activeTab = 0
	for i, id := range dl.centerTabs {
		if id == s.ActiveID {
			dl.activeTab = i
			break
		}
	}
	dl.scale = s.Scale
	return nil
}

func validateDock(s DockSnapshot) error {
	if s.Scale <= 0 {
		return fmt.Errorf("dock: scale")
	}
	seen := make(map[string]DockPanelState, len(s.Panels))
	for _, p := range s.Panels {
		if p.ID == "" {
			return fmt.Errorf("dock: empty id")
		}
		if _, ok := seen[p.ID]; ok {
			return fmt.Errorf("dock: duplicate id")
		}
		if p.Position > DockFloat {
			return fmt.Errorf("dock: position")
		}
		if p.DockSize < 0 || p.FloatW < 0 || p.FloatH < 0 {
			return fmt.Errorf("dock: size")
		}
		seen[p.ID] = p
	}
	center := make(map[string]int, len(s.Center))
	for _, id := range s.Center {
		p, ok := seen[id]
		if !ok || p.Position != DockCenter {
			return fmt.Errorf("dock: center")
		}
		center[id]++
		if center[id] > 1 {
			return fmt.Errorf("dock: center")
		}
	}
	for _, p := range s.Panels {
		if p.Position == DockCenter && center[p.ID] != 1 {
			return fmt.Errorf("dock: center")
		}
	}
	if s.ActiveID != "" && center[s.ActiveID] != 1 {
		return fmt.Errorf("dock: active")
	}
	return nil
}
