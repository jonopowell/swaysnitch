# Testing Waysnitch Keyboard Capture

## Running the Logger

```bash
./waysnitch &
```

## Sending Test Keyboard Events

Once the application is running, you can send keyboard events in two ways:

### Method 1: Using the FIFO

The application creates a FIFO at `/tmp/waysnitch-keys`. Send events like this:

```bash
# Send a key press (W key, scancode 17)
echo "keycode:17,state:1" > /tmp/waysnitch-keys

# Send a key release
echo "keycode:17,state:0" > /tmp/waysnitch-keys

# Send multiple keys
echo "keycode:42,state:1" > /tmp/waysnitch-keys  # Shift press
echo "keycode:17,state:1" > /tmp/waysnitch-keys  # W press
echo "keycode:17,state:0" > /tmp/waysnitch-keys  # W release
echo "keycode:42,state:0" > /tmp/waysnitch-keys  # Shift release
```

### Method 2: Scripting Examples

```bash
# Simulate typing "hello"
for code in 30 18 35 35 39; do
    echo "keycode:$code,state:1" > /tmp/waysnitch-keys
    sleep 0.1
    echo "keycode:$code,state:0" > /tmp/waysnitch-keys
    sleep 0.1
done
```

## Scancode Reference

Common keycodes (from Linux input subsystem):
- 30 = A
- 17 = W
- 31 = S
- 32 = D
- 42 = Shift (left)
- 54 = Shift (right)
- 56 = Alt
- 57 = Space
- 28 = Enter
- 1 = Esc

For a full list, see the `scanCodeToChar` mapping in main.go.

## Integration with Real Wayland

In a real Wayland environment with full input support, the application will automatically:
- Detect and connect to the Wayland compositor
- Monitor wl_keyboard protocol events
- Capture raw keyboard input events

The FIFO method serves as a fallback for testing and restricted environments.
