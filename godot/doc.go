// Package godot provides integration between the GoUI library and the Godot
// game engine via GDExtension.
//
// # Architecture
//
// The integration renders UI to an RGBA pixel buffer using a software rasterizer.
// Godot reads these pixels into an ImageTexture for display on a TextureRect or
// Sprite2D node. This avoids GPU interop complexity and works on all platforms.
//
//	┌─────────────────────────────────────────────┐
//	│  Godot (GDScript / C#)                      │
//	│                                             │
//	│   TextureRect ← ImageTexture ← RGBA pixels │
//	│        ↑              ↑                     │
//	│   InputEvent ─→  godot.UI.InjectEvent()     │
//	│                  godot.UI.Frame(dt)          │
//	│                  godot.UI.Pixels() ──────┐  │
//	└──────────────────────────────────────────┘  │
//	                                               │
//	┌──────────────────────────────────────────────┘
//	│  Go (c-shared / GDExtension)
//	│
//	│   godot.UI → widget tree → CommandBuffer
//	│              → SoftwareBackend.Submit()
//	│              → RGBA pixel buffer
//	└─────────────────────────────────────────────┘
//
// # Usage from GDScript
//
// Build the Go code as a shared library:
//
//	CGO_ENABLED=1 go build -buildmode=c-shared -o goui.dll ./godot/cmd/gdext
//
// Register as GDExtension in Godot project (goui.gdextension):
//
//	[configuration]
//	entry_symbol = "goui_init"
//	[libraries]
//	windows.x86_64 = "res://bin/goui.dll"
//	linux.x86_64   = "res://bin/goui.so"
//	macos.arm64    = "res://bin/goui.dylib"
//
// Then in GDScript:
//
//	var ui = GoUI.new()
//	ui.load_html('<div style="color:white">Hello Godot</div>')
//	# In _process(delta):
//	ui.frame(delta)
//	$TextureRect.texture.update(ui.get_image())
package godot
