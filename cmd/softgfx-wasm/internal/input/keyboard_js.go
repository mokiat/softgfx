// +build js

package input

import (
	"fmt"
	"sync"
	"syscall/js"
)

// NewKeyboard creates a new Keyboard instance that will track
// keyboard events on the HTML element with the specified elementID.
// If elementID is blank, then Keyboard will attach to the HTML document
// element.
// Once the Keyboard is no longer needed, the Destroy method should be
// called to unsubscribe from the HTML element and release allocated resources.
func NewKeyboard(elementID string) (*Keyboard, error) {
	htmlDocument := js.Global().Get("document")
	if htmlDocument.IsUndefined() {
		return nil, fmt.Errorf("could not locate document element")
	}

	htmlTargetElement := htmlDocument
	if elementID != "" {
		htmlTargetElement = htmlDocument.Call("getElementById", elementID)
		if htmlTargetElement.IsUndefined() {
			return nil, fmt.Errorf("could not locate element with id: %s", elementID)
		}
	}

	keyboard := &Keyboard{
		htmlElement: htmlTargetElement,
		keymap:      make(map[KeyName]struct{}),
	}
	keyboard.subscribeKeyEvents()
	return keyboard, nil
}

// Keyboard tracks keyboard events on a given HTML element.
type Keyboard struct {
	htmlElement js.Value

	keyLock sync.Mutex
	keymap  map[KeyName]struct{}

	keydownCallback js.Func
	keyupCallback   js.Func
}

// IsKeyPressed returns whether the key with the specified name is
// currently pressed or not.
func (k *Keyboard) IsKeyPressed(name KeyName) bool {
	k.keyLock.Lock()
	defer k.keyLock.Unlock()
	_, pressed := k.keymap[name]
	return pressed
}

// Destroy releases allocated resources by unsubscribing from key events
func (k *Keyboard) Destroy() {
	k.unsubscribeKeyEvents()
}

func (k *Keyboard) onKeyDown(this js.Value, args []js.Value) interface{} {
	event := args[0]
	event.Call("preventDefault")
	name := KeyName(event.Get("key").String())

	k.keyLock.Lock()
	defer k.keyLock.Unlock()
	k.keymap[name] = struct{}{}
	return true
}

func (k *Keyboard) onKeyUp(this js.Value, args []js.Value) interface{} {
	event := args[0]
	event.Call("preventDefault")
	name := KeyName(event.Get("key").String())

	k.keyLock.Lock()
	defer k.keyLock.Unlock()
	delete(k.keymap, name)
	return true
}

func (k *Keyboard) subscribeKeyEvents() {
	k.keyupCallback = js.FuncOf(k.onKeyUp)
	k.htmlElement.Call("addEventListener", "keyup", k.keyupCallback)

	k.keydownCallback = js.FuncOf(k.onKeyDown)
	k.htmlElement.Call("addEventListener", "keydown", k.keydownCallback)
}

func (k *Keyboard) unsubscribeKeyEvents() {
	k.htmlElement.Call("removeEventListener", "keydown", k.keydownCallback)
	k.keydownCallback.Release()

	k.htmlElement.Call("removeEventListener", "keyup", k.keyupCallback)
	k.keyupCallback.Release()
}
