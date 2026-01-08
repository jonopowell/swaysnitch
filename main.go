package main

/*
#cgo pkg-config: wayland-client
#include <wayland-client.h>
#include <wayland-client-protocol.h>
#include <unistd.h>
#include <string.h>
#include <stdlib.h>
#include <linux/input-event-codes.h>
#include <fcntl.h>
#include <sys/mman.h>

#include "xdg-shell-client-protocol.h"

// Forward declarations for listener callbacks (CGo exported functions)
void registry_global(void *data, struct wl_registry *registry, uint32_t name, char *interface, uint32_t version);
void registry_global_remove(void *data, struct wl_registry *registry, uint32_t name);

void seat_capabilities(void *data, struct wl_seat *wl_seat, uint32_t capabilities);
void seat_name(void *data, struct wl_seat *wl_seat, char *name);

// Wrappers (implemented in listeners.c)

// Input callbacks
void pointer_enter(void *data, struct wl_pointer *wl_pointer, uint32_t serial, struct wl_surface *surface, wl_fixed_t surface_x, wl_fixed_t surface_y);
void pointer_leave(void *data, struct wl_pointer *wl_pointer, uint32_t serial, struct wl_surface *surface);
void pointer_motion(void *data, struct wl_pointer *wl_pointer, uint32_t time, wl_fixed_t surface_x, wl_fixed_t surface_y);
void pointer_button(void *data, struct wl_pointer *wl_pointer, uint32_t serial, uint32_t time, uint32_t button, uint32_t state);
void pointer_axis(void *data, struct wl_pointer *wl_pointer, uint32_t time, uint32_t axis, wl_fixed_t value);

void keyboard_keymap(void *data, struct wl_keyboard *wl_keyboard, uint32_t format, int32_t fd, uint32_t size);
void keyboard_enter(void *data, struct wl_keyboard *wl_keyboard, uint32_t serial, struct wl_surface *surface, struct wl_array *keys);
void keyboard_leave(void *data, struct wl_keyboard *wl_keyboard, uint32_t serial, struct wl_surface *surface);
void keyboard_key(void *data, struct wl_keyboard *wl_keyboard, uint32_t serial, uint32_t time, uint32_t key, uint32_t state);
void keyboard_modifiers(void *data, struct wl_keyboard *wl_keyboard, uint32_t serial, uint32_t mods_depressed, uint32_t mods_latched, uint32_t mods_locked, uint32_t group);
void keyboard_repeat_info(void *data, struct wl_keyboard *wl_keyboard, int32_t rate, int32_t delay);

void touch_down(void *data, struct wl_touch *wl_touch, uint32_t serial, uint32_t time, struct wl_surface *surface, int32_t id, wl_fixed_t x, wl_fixed_t y);
void touch_up(void *data, struct wl_touch *wl_touch, uint32_t serial, uint32_t time, int32_t id);
void touch_motion(void *data, struct wl_touch *wl_touch, uint32_t time, int32_t id, wl_fixed_t x, wl_fixed_t y);
void touch_frame(void *data, struct wl_touch *wl_touch);
void touch_cancel(void *data, struct wl_touch *wl_touch);

// XDG Shell callbacks
void xdg_wm_base_ping(void *data, struct xdg_wm_base *xdg_wm_base, uint32_t serial);

void xdg_surface_configure(void *data, struct xdg_surface *xdg_surface, uint32_t serial);

void xdg_toplevel_configure(void *data, struct xdg_toplevel *xdg_toplevel, int32_t width, int32_t height, struct wl_array *states);
void xdg_toplevel_close(void *data, struct xdg_toplevel *xdg_toplevel);

// Listener structs (implemented in listeners.c)
extern const struct wl_registry_listener registry_listener;
extern const struct wl_seat_listener seat_listener;
extern const struct wl_pointer_listener pointer_listener;
extern const struct wl_keyboard_listener keyboard_listener;
extern const struct wl_touch_listener touch_listener;
extern const struct xdg_wm_base_listener xdg_wm_base_listener;
extern const struct xdg_surface_listener xdg_surface_listener;
extern const struct xdg_toplevel_listener xdg_toplevel_listener;

*/
import "C"

import (
	"fmt"
	"log"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var appState *AppState

type Seat struct {
	seat     *C.struct_wl_seat
	pointer  *C.struct_wl_pointer
	keyboard *C.struct_wl_keyboard
	touch    *C.struct_wl_touch
	name     string
}

type AppState struct {
	display    *C.struct_wl_display
	registry   *C.struct_wl_registry
	compositor *C.struct_wl_compositor
	shm        *C.struct_wl_shm
	
	seats      []*Seat

	xdgWmBase   *C.struct_xdg_wm_base
	surface     *C.struct_wl_surface
	xdgSurface  *C.struct_xdg_surface
	xdgToplevel *C.struct_xdg_toplevel

	width      int32
	height     int32
	closed     bool
	configured bool

	buffer   *C.struct_wl_buffer
	shm_data []byte

	events []string
	mu     sync.Mutex
}

func init() {
	appState = &AppState{
		width:  800,
		height: 600,
		events: make([]string, 0),
		seats:  make([]*Seat, 0),
	}
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func AddEvent(msg string) {
	appState.mu.Lock()

	ts := time.Now().Format("15:04:05.000")
	line := fmt.Sprintf("[%s] %s", ts, msg)

	appState.events = append(appState.events, line)
	if len(appState.events) > 30 {
		appState.events = appState.events[len(appState.events)-30:]
	}
	isConfigured := appState.configured
	appState.mu.Unlock() // Unlock before Draw to avoid deadlock

	// If configured, redraw
	if isConfigured {
		Draw()
	}
}

func createShmBuffer(width, height int32) {
	// Simple SHM buffer creation
	stride := width * 4
	size := stride * height

	fd, err := createAnonymousFile(int64(size))
	if err != nil {
		log.Fatalf("creating anonymous file failed: %v", err)
	}

	data, err := syscall.Mmap(int(fd), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("mmap failed: %v", err)
	}
	appState.shm_data = data

	pool := C.wl_shm_create_pool(appState.shm, C.int32_t(fd), C.int32_t(size))
	appState.buffer = C.wl_shm_pool_create_buffer(pool, 0, C.int32_t(width), C.int32_t(height), C.int32_t(stride), C.WL_SHM_FORMAT_ARGB8888)
	C.wl_shm_pool_destroy(pool)
	syscall.Close(int(fd))
}

func createAnonymousFile(size int64) (int, error) {
	name := "waysnitch-shm"
	fd, _, errno := syscall.Syscall(syscall.SYS_MEMFD_CREATE, uintptr(unsafe.Pointer(syscall.StringBytePtr(name))), 0, 0)
	if errno != 0 {
		return 0, errno
	}
	err := syscall.Ftruncate(int(fd), size)
	if err != nil {
		syscall.Close(int(fd))
		return 0, err
	}
	return int(fd), nil
}

func Draw() {
	if appState.shm_data == nil || appState.surface == nil {
		return
	}

	w := int(appState.width)
	h := int(appState.height)

	// Clear
	for i := 0; i < len(appState.shm_data); i += 4 {
		appState.shm_data[i] = 0
		appState.shm_data[i+1] = 0
		appState.shm_data[i+2] = 0
		appState.shm_data[i+3] = 255
	}

	appState.mu.Lock()
	events := make([]string, len(appState.events))
	copy(events, appState.events)
	appState.mu.Unlock()

	y := 20
	DrawString(appState.shm_data, w, "Waysnitch (Pure Wayland/XDG Shell)", 10, y, 0xFF00FF00)
	y += 20

	for _, evt := range events {
		DrawString(appState.shm_data, w, evt, 10, y, 0xFFFFFFFF)
		y += 15
	}

	C.wl_surface_attach(appState.surface, appState.buffer, 0, 0)
	C.wl_surface_damage(appState.surface, 0, 0, C.int32_t(w), C.int32_t(h))
	C.wl_surface_commit(appState.surface)
}

// ----------------------------------------------------------------------------
// Exported C Callbacks
// ----------------------------------------------------------------------------

// Helper to get monotonic time for logging
func now() string {
	return time.Now().Format("15:04:05.000")
}

//export registry_global
func registry_global(data unsafe.Pointer, registry *C.struct_wl_registry, name C.uint32_t, interfaceName *C.char, version C.uint32_t) {
	ifName := C.GoString(interfaceName)
	log.Printf("Registry found: %s v%d", ifName, version)

	if ifName == "wl_compositor" {
		appState.compositor = (*C.struct_wl_compositor)(C.wl_registry_bind(registry, name, &C.wl_compositor_interface, 1))
	} else if ifName == "wl_shm" {
		appState.shm = (*C.struct_wl_shm)(C.wl_registry_bind(registry, name, &C.wl_shm_interface, 1))
	} else if ifName == "xdg_wm_base" {
		appState.xdgWmBase = (*C.struct_xdg_wm_base)(C.wl_registry_bind(registry, name, &C.xdg_wm_base_interface, 1))
		C.xdg_wm_base_add_listener(appState.xdgWmBase, &C.xdg_wm_base_listener, nil)
	} else if ifName == "wl_seat" {
		seatPtr := (*C.struct_wl_seat)(C.wl_registry_bind(registry, name, &C.wl_seat_interface, 4)) // Bind v4 just in case
		newSeat := &Seat{seat: seatPtr}
		appState.seats = append(appState.seats, newSeat)
		C.wl_seat_add_listener(seatPtr, &C.seat_listener, nil)
	}
}

//export registry_global_remove
func registry_global_remove(data unsafe.Pointer, registry *C.struct_wl_registry, name C.uint32_t) {}

//export seat_capabilities
func seat_capabilities(data unsafe.Pointer, wl_seat *C.struct_wl_seat, capabilities C.uint32_t) {
	log.Printf("Seat capabilities: %d", capabilities)

	// Find the seat
	var currentSeat *Seat
	for _, s := range appState.seats {
		if s.seat == wl_seat {
			currentSeat = s
			break
		}
	}
	if currentSeat == nil {
		log.Println("Error: Seat capabilities event for unknown seat")
		return
	}

	if (capabilities & C.WL_SEAT_CAPABILITY_POINTER) != 0 {
		log.Println("Seat has Pointer")
		currentSeat.pointer = C.wl_seat_get_pointer(wl_seat)
		C.wl_pointer_add_listener(currentSeat.pointer, &C.pointer_listener, nil)
	}
	if (capabilities & C.WL_SEAT_CAPABILITY_KEYBOARD) != 0 {
		log.Println("Seat has Keyboard")
		currentSeat.keyboard = C.wl_seat_get_keyboard(wl_seat)
		C.wl_keyboard_add_listener(currentSeat.keyboard, &C.keyboard_listener, nil)
	}
	if (capabilities & C.WL_SEAT_CAPABILITY_TOUCH) != 0 {
		log.Println("Seat has Touch")
		currentSeat.touch = C.wl_seat_get_touch(wl_seat)
		C.wl_touch_add_listener(currentSeat.touch, &C.touch_listener, nil)
	}
}

//export seat_name
func seat_name(data unsafe.Pointer, wl_seat *C.struct_wl_seat, name *C.char) {
	seatName := C.GoString(name)
	log.Printf("Seat name: %s", seatName)

	var currentSeat *Seat
	for _, s := range appState.seats {
		if s.seat == wl_seat {
			currentSeat = s
			break
		}
	}
	if currentSeat != nil {
		currentSeat.name = seatName
	}
}

// -- XDG Shell --

//export xdg_wm_base_ping
func xdg_wm_base_ping(data unsafe.Pointer, base *C.struct_xdg_wm_base, serial C.uint32_t) {
	C.xdg_wm_base_pong(base, serial)
}

//export xdg_surface_configure
func xdg_surface_configure(data unsafe.Pointer, xdg_surface *C.struct_xdg_surface, serial C.uint32_t) {
	C.xdg_surface_ack_configure(xdg_surface, serial)

	if appState.buffer == nil {
		createShmBuffer(appState.width, appState.height)
	}

	log.Println("Window Configured (Ready to Draw)")
	appState.configured = true
	Draw()
}

//export xdg_toplevel_configure
func xdg_toplevel_configure(data unsafe.Pointer, toplevel *C.struct_xdg_toplevel, width, height C.int32_t, states *C.struct_wl_array) {
	if width > 0 && height > 0 {
		appState.width = int32(width)
		appState.height = int32(height)
		// Recreate buffer? Simplified: no, just careful drawing or resize
		// For proper window resizing, we should destroy old buffer and create new one
		// But let's stick to fixed size for MVP
	}
}

//export xdg_toplevel_close
func xdg_toplevel_close(data unsafe.Pointer, toplevel *C.struct_xdg_toplevel) {
	appState.closed = true
}

// -- Inputs --

//export pointer_enter
func pointer_enter(data unsafe.Pointer, wl_pointer *C.struct_wl_pointer, serial C.uint32_t, surface *C.struct_wl_surface, surface_x C.wl_fixed_t, surface_y C.wl_fixed_t) {
	AddEvent("Pointer Enter")
}

//export pointer_leave
func pointer_leave(data unsafe.Pointer, wl_pointer *C.struct_wl_pointer, serial C.uint32_t, surface *C.struct_wl_surface) {
	AddEvent("Pointer Leave")
}

//export pointer_motion
func pointer_motion(data unsafe.Pointer, wl_pointer *C.struct_wl_pointer, time C.uint32_t, surface_x C.wl_fixed_t, surface_y C.wl_fixed_t) {
}

//export pointer_button
func pointer_button(data unsafe.Pointer, wl_pointer *C.struct_wl_pointer, serial C.uint32_t, time C.uint32_t, button C.uint32_t, state C.uint32_t) {
	s := "Release"
	if state == 1 {
		s = "Press"
	}
	AddEvent(fmt.Sprintf("Button %d %s", button, s))
}

//export pointer_axis
func pointer_axis(data unsafe.Pointer, wl_pointer *C.struct_wl_pointer, time C.uint32_t, axis C.uint32_t, value C.wl_fixed_t) {
	AddEvent("Scroll")
}

//export keyboard_keymap
func keyboard_keymap(data unsafe.Pointer, wl_keyboard *C.struct_wl_keyboard, format C.uint32_t, fd C.int32_t, size C.uint32_t) {
	syscall.Close(int(fd))
}

//export keyboard_enter
func keyboard_enter(data unsafe.Pointer, wl_keyboard *C.struct_wl_keyboard, serial C.uint32_t, surface *C.struct_wl_surface, keys *C.struct_wl_array) {
	log.Println("EVENT: Keyboard Focus Gained")
	AddEvent("Keyboard Focus Gained")
}

//export keyboard_leave
func keyboard_leave(data unsafe.Pointer, wl_keyboard *C.struct_wl_keyboard, serial C.uint32_t, surface *C.struct_wl_surface) {
	log.Println("EVENT: Keyboard Focus Lost")
	AddEvent("Keyboard Focus Lost")
}

//export keyboard_key
func keyboard_key(data unsafe.Pointer, wl_keyboard *C.struct_wl_keyboard, serial C.uint32_t, time C.uint32_t, key C.uint32_t, state C.uint32_t) {
	log.Printf("Keyboard event: key=%d state=%d", key, state) // Debug log
	s := "Up"
	if state == 1 {
		s = "Down"
	}
	kc := int(key)
	keyName := ""
	if name, ok := scanCodeToChar[kc]; ok {
		keyName = name
	} else {
		keyName = fmt.Sprintf("KEY_%d", kc)
	}
	AddEvent(fmt.Sprintf("%s %s (%d)", s, keyName, kc))
}

//export keyboard_modifiers
func keyboard_modifiers(data unsafe.Pointer, wl_keyboard *C.struct_wl_keyboard, serial C.uint32_t, mods_depressed C.uint32_t, mods_latched C.uint32_t, mods_locked C.uint32_t, group C.uint32_t) {
}

//export keyboard_repeat_info
func keyboard_repeat_info(data unsafe.Pointer, wl_keyboard *C.struct_wl_keyboard, rate C.int32_t, delay C.int32_t) {
}

//export touch_down
func touch_down(data unsafe.Pointer, wl_touch *C.struct_wl_touch, serial C.uint32_t, time C.uint32_t, surface *C.struct_wl_surface, id C.int32_t, x C.wl_fixed_t, y C.wl_fixed_t) {
	AddEvent(fmt.Sprintf("Touch Down %d", id))
}

//export touch_up
func touch_up(data unsafe.Pointer, wl_touch *C.struct_wl_touch, serial C.uint32_t, time C.uint32_t, id C.int32_t) {
	AddEvent(fmt.Sprintf("Touch Up %d", id))
}

//export touch_motion
func touch_motion(data unsafe.Pointer, wl_touch *C.struct_wl_touch, time C.uint32_t, id C.int32_t, x C.wl_fixed_t, y C.wl_fixed_t) {
}

//export touch_frame
func touch_frame(data unsafe.Pointer, wl_touch *C.struct_wl_touch) {}

//export touch_cancel
func touch_cancel(data unsafe.Pointer, wl_touch *C.struct_wl_touch) { AddEvent("Touch Cancel") }

// Key map
var scanCodeToChar = map[int]string{
	1: "Esc", 30: "A", 31: "S", 32: "D", 33: "F", 34: "G", 35: "H", 36: "J", 37: "K", 38: "L",
}

func main() {
	appState.display = C.wl_display_connect(nil)
	if appState.display == nil {
		log.Fatal("failed to connect to display")
	}
	defer C.wl_display_disconnect(appState.display)

	appState.registry = C.wl_display_get_registry(appState.display)
	C.wl_registry_add_listener(appState.registry, &C.registry_listener, nil)

	C.wl_display_roundtrip(appState.display)

	if appState.compositor == nil || appState.shm == nil || appState.xdgWmBase == nil {
		log.Fatal("Missing required Wayland globals (compositor, shm, or xdg_wm_base)")
	}

	appState.surface = C.wl_compositor_create_surface(appState.compositor)
	appState.xdgSurface = C.xdg_wm_base_get_xdg_surface(appState.xdgWmBase, appState.surface)
	C.xdg_surface_add_listener(appState.xdgSurface, &C.xdg_surface_listener, nil)

	appState.xdgToplevel = C.xdg_surface_get_toplevel(appState.xdgSurface)
	C.xdg_toplevel_add_listener(appState.xdgToplevel, &C.xdg_toplevel_listener, nil)
	C.xdg_toplevel_set_title(appState.xdgToplevel, C.CString("Waysnitch"))

	C.wl_surface_commit(appState.surface)

	log.Println("Waysnitch started (XDG Shell). Press keys...")

	for !appState.closed {
		if C.wl_display_dispatch(appState.display) == -1 {
			break
		}
	}
}
