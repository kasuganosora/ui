## GoUI Panel — renders a Go UI instance onto a TextureRect.
##
## Usage:
##   1. Build the shared library:
##        CGO_ENABLED=1 go build -buildmode=c-shared -o goui.dll ./godot/cmd/gdext
##   2. Copy goui.dll to res://bin/ in your Godot project.
##   3. Add goui.gdextension to your project root.
##   4. Attach this script to a TextureRect node.
##
## Since GDExtension with Go requires a C bridge, this example shows the
## calling convention. Use GDNative or a C wrapper to call these functions.
##
## The Go shared library exports these C functions:
##   goui_create(width, height, dpi_scale) → handle
##   goui_destroy(handle)
##   goui_frame(handle, delta)
##   goui_resize(handle, width, height)
##   goui_pixels(handle, &len) → *byte
##   goui_framebuffer_size(handle, &w, &h)
##   goui_inject_mouse_move(handle, x, y)
##   goui_inject_mouse_click(handle, x, y, button)
##   goui_inject_scroll(handle, x, y, dx, dy)
##   goui_inject_key_down(handle, key, modifiers)
##   goui_inject_key_up(handle, key, modifiers)
##   goui_inject_char(handle, unicode_char)

extends TextureRect

# UI handle (obtained from goui_create via GDNative bridge)
var _handle: int = 0
var _image: Image
var _texture: ImageTexture

func _ready() -> void:
	# Create UI instance matching this node's size
	var sz = size
	# _handle = GoUI.create(int(sz.x), int(sz.y), 1.0)

	_image = Image.create(int(sz.x), int(sz.y), false, Image.FORMAT_RGBA8)
	_texture = ImageTexture.create_from_image(_image)
	texture = _texture

func _process(delta: float) -> void:
	if _handle == 0:
		return

	# Process one frame
	# GoUI.frame(_handle, delta)

	# Get pixel data and update texture
	# var pixels: PackedByteArray = GoUI.get_pixels(_handle)
	# _image.set_data(int(size.x), int(size.y), false, Image.FORMAT_RGBA8, pixels)
	# _texture.update(_image)

func _gui_input(event: InputEvent) -> void:
	if _handle == 0:
		return

	if event is InputEventMouseMotion:
		var e = event as InputEventMouseMotion
		pass # GoUI.inject_mouse_move(_handle, e.position.x, e.position.y)

	elif event is InputEventMouseButton:
		var e = event as InputEventMouseButton
		if e.pressed:
			pass # GoUI.inject_mouse_click(_handle, e.position.x, e.position.y, e.button_index)

	elif event is InputEventKey:
		var e = event as InputEventKey
		var mods = 0
		if e.ctrl_pressed: mods |= 1
		if e.shift_pressed: mods |= 2
		if e.alt_pressed: mods |= 4
		if e.meta_pressed: mods |= 8

		if e.pressed:
			pass # GoUI.inject_key_down(_handle, e.keycode, mods)
			if e.unicode > 0:
				pass # GoUI.inject_char(_handle, e.unicode)
		else:
			pass # GoUI.inject_key_up(_handle, e.keycode, mods)

func _notification(what: int) -> void:
	if what == NOTIFICATION_RESIZED:
		if _handle != 0:
			pass # GoUI.resize(_handle, int(size.x), int(size.y))

func _exit_tree() -> void:
	if _handle != 0:
		pass # GoUI.destroy(_handle)
		_handle = 0
