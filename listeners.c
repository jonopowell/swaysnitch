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
