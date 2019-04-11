package main

import (
	_ "image/png"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"storj.io/snoboard/graphics"
)

const (
	windowWidth  = 1024
	windowHeight = 768
	speed        = 30
)

// Scene represents the root game scene. The scene references graphic resources, objects and game state.
type Scene struct {
	Window             *pixelgl.Window
	LastFrameTime      time.Time
	TimeSinceLastFrame float64
	CameraPosition     pixel.Vec
	Background         *Object
	Player             *Object
	Obstacles          []*Object // TODO: Implement the obstacles in the scene.
}

// Object represents an item in the game (player, obstacle, etc...)
type Object struct {
	position pixel.Vec
	velocity pixel.Vec
	sprite   *pixel.Sprite
}

func main() {
	pixelgl.Run(renderLoop)
}

func renderLoop() {
	scene := initializeScene()

	for !scene.Window.Closed() {
		scene.TimeSinceLastFrame = time.Since(scene.LastFrameTime).Seconds()
		scene.LastFrameTime = time.Now()

		// Call the render pipeline.
		processInput(scene)
		updateState(scene)
		render(scene)
	}
}

// processInput is where we process any input events from the keyboard.
func processInput(scene *Scene) {
	newVelX := float64(0)
	if scene.Window.Pressed(pixelgl.KeyLeft) {
		newVelX -= speed
	}
	if scene.Window.Pressed(pixelgl.KeyRight) {
		newVelX += speed
	}

	scene.Player.velocity = pixel.V(newVelX, scene.Player.velocity.Y)

	newX := scene.Player.position.X + scene.Player.velocity.X*scene.TimeSinceLastFrame
	newY := scene.Player.position.Y + scene.Player.velocity.Y*scene.TimeSinceLastFrame
	scene.Player.position = pixel.V(newX, newY)

	camPosX := scene.Player.position.X - scene.Window.Bounds().Center().X
	camPosY := scene.Player.position.Y - scene.Window.Bounds().Center().Y
	scene.CameraPosition = pixel.V(camPosX, camPosY)
}

// updateState is where we update any game state.
// Maybe update scores, object states that aren't related to input.
func updateState(scene *Scene) {

}

// render is where we render graphics after all the input and game state has been processed.
func render(scene *Scene) {
	cameraMatrix := pixel.IM.Moved(scene.CameraPosition.Scaled(-1))
	scene.Window.SetMatrix(cameraMatrix)

	scene.Window.Clear(colornames.Blueviolet)
	scene.Background.sprite.Draw(scene.Window, pixel.IM.Moved(scene.Background.position))
	scene.Player.sprite.Draw(scene.Window, pixel.IM.Moved(scene.Player.position))

	scene.Window.Update()
}

func initializeScene() *Scene {
	scene := &Scene{}

	// Create the render window.
	cfg := pixelgl.WindowConfig{
		Title:  "SNOboard",
		Bounds: pixel.R(0, 0, windowWidth, windowHeight),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	scene.Window = win

	// Create the player object and assets.
	playerSprite, err := graphics.LoadPicture("gopher-happy.png")
	if err != nil {
		panic(err)
	}

	player := &Object{
		position: win.Bounds().Center(),
		velocity: pixel.V(0, -speed),
		sprite:   pixel.NewSprite(playerSprite, playerSprite.Bounds()),
	}
	scene.Player = player

	// Create background.
	backgroundSprite, err := graphics.LoadPicture("gopher-happy.png")
	if err != nil {
		panic(err)
	}
	background := &Object{
		position: win.Bounds().Center(),
		velocity: pixel.V(0, 0),
		sprite:   pixel.NewSprite(backgroundSprite, backgroundSprite.Bounds()),
	}
	scene.Background = background

	scene.LastFrameTime = time.Now()
	return scene
}
