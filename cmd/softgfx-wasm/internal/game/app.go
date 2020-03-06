package game

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sync"

	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/bsp"
	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/graphics"
	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/input"
	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/metrics"
	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/scene"
	"github.com/mokiat/softgfx/internal/data"
)

const (
	turnSpeed = float32(120.0)
	walkSpeed = float32(125.0)
	runSpeed  = float32(200.0)
	jumpSpeed = float32(125.0)
	lookSpeed = float32(1.0)
)

func NewApplication(keyboard *input.Keyboard, plotter *graphics.Plotter) *Application {
	sceneRenderer := scene.NewRenderer(plotter)
	bspRenderer := bsp.NewRenderer(sceneRenderer)

	return &Application{
		keyboard:    keyboard,
		plotter:     plotter,
		bspRenderer: bspRenderer,

		initializedMU: &sync.Mutex{},
		initialized:   false,
		camera:        scene.NewCamera(),
	}
}

type Application struct {
	keyboard       *input.Keyboard
	plotter        *graphics.Plotter
	bspRenderer    *bsp.Renderer
	renderDuration metrics.Duration

	initializedMU *sync.Mutex
	initialized   bool
	camera        *scene.Camera
	rootWall      *bsp.Wall
}

func (a *Application) Init(level string) {
	a.initializedMU.Lock()
	defer a.initializedMU.Unlock()
	a.initialized = false

	go func() {
		if err := a.initScene(level); err != nil {
			panic(fmt.Errorf("failed to init scene: %w", err))
		}
	}()
}

func (a *Application) OnUpdate(elapsedSeconds float32) {
	a.initializedMU.Lock()
	defer a.initializedMU.Unlock()
	if !a.initialized {
		return
	}

	a.updatePlayer(elapsedSeconds)
	a.renderDuration.Measure(func() {
		a.bspRenderer.Clear()
		a.bspRenderer.RenderBSP(a.rootWall, a.camera)
	})
	a.plotter.Flush()

	a.renderDuration.Print(60)
}

func (a *Application) updatePlayer(elapsedSeconds float32) {
	if a.keyboard.IsKeyPressed(input.KeyNameUp) || a.keyboard.IsKeyPressed(input.KeyName("w")) {
		a.camera.MoveForward(runSpeed * elapsedSeconds)
	}
	if a.keyboard.IsKeyPressed(input.KeyNameDown) || a.keyboard.IsKeyPressed(input.KeyName("s")) {
		a.camera.MoveBackward(runSpeed * elapsedSeconds)
	}
	if a.keyboard.IsKeyPressed(input.KeyName("a")) {
		a.camera.MoveLeft(walkSpeed * elapsedSeconds)
	}
	if a.keyboard.IsKeyPressed(input.KeyName("d")) {
		a.camera.MoveRight(walkSpeed * elapsedSeconds)
	}
	if a.keyboard.IsKeyPressed(input.KeyNameLeft) {
		a.camera.TurnLeft(turnSpeed * elapsedSeconds)
	}
	if a.keyboard.IsKeyPressed(input.KeyNameRight) {
		a.camera.TurnRight(turnSpeed * elapsedSeconds)
	}
	if a.keyboard.IsKeyPressed(input.KeyNameSpace) {
		a.camera.MoveUp(jumpSpeed * elapsedSeconds)
	}
	if a.keyboard.IsKeyPressed(input.KeyNameShift) {
		a.camera.MoveDown(jumpSpeed * elapsedSeconds)
	}
	if a.keyboard.IsKeyPressed(input.KeyName("q")) {
		a.camera.LookUp(lookSpeed * elapsedSeconds)
	}
	if a.keyboard.IsKeyPressed(input.KeyName("e")) {
		a.camera.LookDown(lookSpeed * elapsedSeconds)
	}
}

func (a *Application) initScene(levelName string) error {
	level, err := fetchLevel(levelName)
	if err != nil {
		return fmt.Errorf("failed to fetch level %q: %w", levelName, err)
	}

	textures := make([]*graphics.Texture, len(level.Textures))
	for i, textureName := range level.Textures {
		texture, err := fetchTexture(textureName)
		if err != nil {
			return fmt.Errorf("failed ot fetch texture %q: %w", textureName, err)
		}
		textures[i] = convertTexture(texture)
	}

	getTexture := func(index int) *graphics.Texture {
		if index < 0 || index >= len(textures) {
			return nil
		}
		return textures[index]
	}

	walls := make([]*bsp.Wall, len(level.Walls))
	for i, levelWall := range level.Walls {
		deltaX := float64(levelWall.RightEdgeX - levelWall.LeftEdgeX)
		deltaZ := float64(levelWall.RightEdgeZ - levelWall.LeftEdgeZ)
		wall := &bsp.Wall{
			LeftEdgeX:  levelWall.LeftEdgeX,
			LeftEdgeZ:  levelWall.LeftEdgeZ,
			RightEdgeX: levelWall.RightEdgeX,
			RightEdgeZ: levelWall.RightEdgeZ,
			Length:     float32(math.Sqrt(deltaX*deltaX + deltaZ*deltaZ)),
		}
		if levelWall.Ceiling != nil {
			wall.Ceiling = &bsp.Extrusion{
				Top:          levelWall.Ceiling.Top,
				Bottom:       levelWall.Ceiling.Bottom,
				OuterTexture: getTexture(levelWall.Ceiling.OuterTexture),
				FaceTexture:  getTexture(levelWall.Ceiling.FaceTexture),
				InnerTexture: getTexture(levelWall.Ceiling.InnerTexture),
			}
		}
		if levelWall.Floor != nil {
			wall.Floor = &bsp.Extrusion{
				Top:          levelWall.Floor.Top,
				Bottom:       levelWall.Floor.Bottom,
				OuterTexture: getTexture(levelWall.Floor.OuterTexture),
				FaceTexture:  getTexture(levelWall.Floor.FaceTexture),
				InnerTexture: getTexture(levelWall.Floor.InnerTexture),
			}
		}
		walls[i] = wall
	}
	for i, levelWall := range level.Walls {
		if frontIndex := levelWall.FrontWall; frontIndex >= 0 {
			walls[i].FrontWall = walls[frontIndex]
		}
		if backIndex := levelWall.BackWall; backIndex >= 0 {
			walls[i].BackWall = walls[backIndex]
		}
	}

	a.initializedMU.Lock()
	defer a.initializedMU.Unlock()
	a.camera.SetPosition(0.0, 0.0, 0.0)
	a.camera.SetRotation(0.0)
	a.rootWall = walls[0]
	a.initialized = true

	return nil
}

func fetchLevel(name string) (data.Level, error) {
	resp, err := http.Get(fmt.Sprintf("web/levels/%s.json", url.PathEscape(name)))
	if err != nil {
		return data.Level{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	level, err := data.LoadLevel(resp.Body)
	if err != nil {
		return data.Level{}, fmt.Errorf("failed to load level: %w", err)
	}
	return level, nil
}

func fetchTexture(name string) (data.Texture, error) {
	resp, err := http.Get(fmt.Sprintf("web/images/%s.png", url.PathEscape(name)))
	if err != nil {
		return data.Texture{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	texture, err := data.LoadTexture(resp.Body)
	if err != nil {
		return data.Texture{}, fmt.Errorf("failed to load texture: %w", err)
	}
	return texture, nil
}

func convertTexture(original data.Texture) *graphics.Texture {
	return &graphics.Texture{
		Width:  original.Width,
		Height: original.Height,
		Texels: original.Texels,
	}
}
