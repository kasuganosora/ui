package accessibility

// Collect reads accessibility metadata from widgets in order.
// Nil entries are skipped.
func Collect(items []Accessible) []Props {
	out := make([]Props, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, item.AccessibleProps())
	}
	return out
}
