// +build !js

package input

func NewKeyboard(elementID string) (*Keyboard, error) {
	return &Keyboard{}, nil
}

type Keyboard struct {
}

func (k *Keyboard) IsKeyPressed(name KeyName) bool {
	return false
}

func (k *Keyboard) Destroy() {
}
