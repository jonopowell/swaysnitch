#include <wayland-client.h>
#include <wayland-client-protocol.h>
#include "xdg-shell-client-protocol.h"
#include "_cgo_export.h" // For Go callback references

// Wrappers to handle const char* vs char* mismatch (if needed, but simpler to just cast in the struct init if compiler allows, or keep wrapper)
// Since we are in a C file, we can just define the functions.

void registry_global_wrapper(void *data, struct wl_registry *registry, uint32_t name, const char *interface, uint32_t version) {
    registry_global(data, registry, name, (char*)interface, version);
}

void seat_name_wrapper(void *data, struct wl_seat *wl_seat, const char *name) {
    seat_name(data, wl_seat, (char*)name);
}

// Listener structs
const struct wl_registry_listener registry_listener = {
	.global = registry_global_wrapper,
	.global_remove = registry_global_remove,
};

const struct wl_seat_listener seat_listener = {
	.capabilities = seat_capabilities,
	.name = seat_name_wrapper,
};

const struct wl_pointer_listener pointer_listener = {
	.enter = pointer_enter,
	.leave = pointer_leave,
	.motion = pointer_motion,
	.button = pointer_button,
	.axis = pointer_axis,
};

const struct wl_keyboard_listener keyboard_listener = {
	.keymap = keyboard_keymap,
	.enter = keyboard_enter,
	.leave = keyboard_leave,
	.key = keyboard_key,
	.modifiers = keyboard_modifiers,
	.repeat_info = keyboard_repeat_info,
};

const struct wl_touch_listener touch_listener = {
	.down = touch_down,
	.up = touch_up,
	.motion = touch_motion,
	.frame = touch_frame,
	.cancel = touch_cancel,
};

// Text Input Listener
#include "text-input-unstable-v1-client-protocol.h"

void text_input_commit_string_wrapper(void *data, struct zwp_text_input_v1 *ti, uint32_t serial, const char *text) {
	if (text) text_input_commit_string(data, ti, serial, (char*)text);
}

void text_input_preedit_string_stub(void *data, struct zwp_text_input_v1 *ti, uint32_t serial, const char *text, const char *commit) {}
void text_input_modifiers_map_stub(void *data, struct zwp_text_input_v1 *ti, struct wl_array *map) {}
void text_input_input_panel_state_stub(void *data, struct zwp_text_input_v1 *ti, uint32_t state) {}
void text_input_preedit_styling_stub(void *data, struct zwp_text_input_v1 *ti, uint32_t index, uint32_t length, uint32_t style) {}
void text_input_preedit_cursor_stub(void *data, struct zwp_text_input_v1 *ti, int32_t index) {}
void text_input_cursor_position_stub(void *data, struct zwp_text_input_v1 *ti, int32_t index, int32_t anchor) {}
void text_input_delete_surrounding_text_stub(void *data, struct zwp_text_input_v1 *ti, int32_t index, uint32_t length) {}
void text_input_keysym_stub(void *data, struct zwp_text_input_v1 *ti, uint32_t serial, uint32_t time, uint32_t sym, uint32_t state, uint32_t modifiers) {}
void text_input_language_stub(void *data, struct zwp_text_input_v1 *ti, uint32_t serial, const char *language) {}
void text_input_text_direction_stub(void *data, struct zwp_text_input_v1 *ti, uint32_t serial, uint32_t direction) {}

const struct zwp_text_input_v1_listener text_input_listener = {
	.enter = text_input_enter,
	.leave = text_input_leave,
	.modifiers_map = text_input_modifiers_map_stub,
	.input_panel_state = text_input_input_panel_state_stub,
	.preedit_string = text_input_preedit_string_stub,
	.preedit_styling = text_input_preedit_styling_stub,
	.preedit_cursor = text_input_preedit_cursor_stub,
	.commit_string = text_input_commit_string_wrapper,
	.cursor_position = text_input_cursor_position_stub,
	.delete_surrounding_text = text_input_delete_surrounding_text_stub,
	.keysym = text_input_keysym_stub,
	.language = text_input_language_stub,
	.text_direction = text_input_text_direction_stub,
};

const struct xdg_wm_base_listener xdg_wm_base_listener = {
	.ping = xdg_wm_base_ping,
};

const struct xdg_surface_listener xdg_surface_listener = {
	.configure = xdg_surface_configure,
};

const struct xdg_toplevel_listener xdg_toplevel_listener = {
	.configure = xdg_toplevel_configure,
	.close = xdg_toplevel_close,
};
