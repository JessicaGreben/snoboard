package main

import (
	"fmt"
	_ "image/png"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"storj.io/snoboard/graphics"
)

const (
	windowWidth  = 1024
	windowHeight = 768
	speed        = 80
)

// Scene represents the root game scene. The scene references graphic resources, objects and game state.
type Scene struct {
	Window                *pixelgl.Window
	LastFrameTime         time.Time
	TimeSinceLastFrame    float64
	TimeSinceLastObstacle float64
	CameraPosition        pixel.Vec
	Background            *Object
	Player                *Object
	Obstacles             []*Object
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
		scene.TimeSinceLastObstacle += scene.TimeSinceLastFrame

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
	player := scene.Player
	backgroundSprite, err := graphics.LoadPicture("pot.png")
	if err != nil {
		panic(err)
	}

	var lastIndex int
	for i, o := range scene.Obstacles {
		if o.position.Y > player.position.Y+float64(400) {
			lastIndex = i
		}
	}
	scene.Obstacles = scene.Obstacles[lastIndex:]

	detectCollisions(scene)

	// If it has been 1 second since last obstacle then create a new one
	if scene.TimeSinceLastObstacle > 1 {
		randX := rand.Intn(2*windowWidth) - windowWidth
		newObj := &Object{
			position: player.position.Add(pixel.V(float64(randX), -400)),
			velocity: pixel.V(0, 0),
			sprite:   pixel.NewSprite(backgroundSprite, backgroundSprite.Bounds()),
		}
		scene.Obstacles = append(scene.Obstacles, newObj)
		scene.TimeSinceLastObstacle = 0
	}
}

func detectCollisions(scene *Scene) {
	for _, obstacle := range scene.Obstacles {
		if intersectRect(scene.Player, obstacle) {
			panic("HIT!!!!!!!!!!!!!!!!!!!!!!!!!")
		}
	}
}

func intersectRect(object1 *Object, object2 *Object) bool {
	object1Right := object1.position.X + object1.sprite.Frame().W()
	object2Right := object2.position.X + object2.sprite.Frame().W()
	object1Bottom := object1.position.Y + object1.sprite.Frame().H()
	object2Bottom := object2.position.Y + object2.sprite.Frame().H()

	collides := !(object2.position.X > object1Right ||
		object2Right < object1.position.X ||
		object2.position.Y > object1Bottom ||
		object2Bottom < object1.position.Y)

	fmt.Printf("[TOP: %4.0f BOTTOM: %4.0f LEFT: %4.0f RIGHT: %4.0f]\t[TOP: %4.0f BOTTOM: %4.0f LEFT: %4.0f RIGHT: %4.0f]\n",
		object1.position.Y,
		object1Bottom,
		object1.position.X,
		object1Right,
		object2.position.Y,
		object2Bottom,
		object2.position.X,
		object2Right)

	return collides
}

// render is where we render graphics after all the input and game state has been processed.
func render(scene *Scene) {
	cameraMatrix := pixel.IM.Moved(scene.CameraPosition.Scaled(-1))
	scene.Window.SetMatrix(cameraMatrix)

	scene.Window.Clear(colornames.Blueviolet)
	scene.Player.sprite.Draw(scene.Window, pixel.IM.Moved(scene.Player.position))
	for _, o := range scene.Obstacles {
		o.sprite.Draw(scene.Window, pixel.IM.Moved(o.position))
	}

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
	scene.LastFrameTime = time.Now()
	return scene
}
