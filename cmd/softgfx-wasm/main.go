package main

import (
	"fmt"
	"time"

	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/game"
	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/graphics"
	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/input"
)

func main() {
	keyboard, err := input.NewKeyboard("game")
	if err != nil {
		panic(fmt.Errorf("could not create keyboard: %s", err))
	}
	defer keyboard.Destroy()

	plotter, err := graphics.NewPlotter("screen")
	if err != nil {
		panic(fmt.Errorf("could not create plotter: %s", err))
	}

	app := game.NewApplication(keyboard, plotter)
	app.Init("castle")

	lastFrameTime := time.Now()
	for range time.Tick(15 * time.Millisecond) {
		currentTime := time.Now()
		elapsedSeconds := currentTime.Sub(lastFrameTime).Seconds()
		lastFrameTime = currentTime

		app.OnUpdate(float32(elapsedSeconds))
	}
}
