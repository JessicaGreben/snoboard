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
	speed        = 300
)

// Scene represents the root game scene. The scene references graphic resources, objects and game state.
type Scene struct {
	Window                *pixelgl.Window
	LastFrameTime         time.Time
	TimeSinceLastFrame    float64
	TimeSinceLastObstacle float64
	CameraPosition        pixel.Vec
	Player                *Object
	Obstacles             []*Object
	Sprites               *Sprites
	Dead                  bool
}

// Object represents an item in the game (player, obstacle, etc...)
type Object struct {
	position pixel.Vec
	velocity pixel.Vec
	sprite   *pixel.Sprite
}

// Sprites are all the images we use
type Sprites struct {
	forward   *pixel.Sprite
	left      *pixel.Sprite
	right     *pixel.Sprite
	server    *pixel.Sprite
	harddrive *pixel.Sprite
	jump      *pixel.Sprite
	jumpleft  *pixel.Sprite
	jumpright *pixel.Sprite
	wipeout   *pixel.Sprite
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
		scene.Player.sprite = scene.Sprites.left
	}
	if scene.Window.Pressed(pixelgl.KeyRight) {
		newVelX += speed
		scene.Player.sprite = scene.Sprites.right
	}
	if !scene.Window.Pressed(pixelgl.KeyRight) && !scene.Window.Pressed(pixelgl.KeyLeft) {
		scene.Player.sprite = scene.Sprites.forward
	}
	if scene.Dead && scene.Window.Pressed(pixelgl.KeySpace) {
		scene.Dead = false
		scene.Player.position = scene.Window.Bounds().Center()
	}

	player := scene.Player
	if scene.Dead {
		player.sprite = scene.Sprites.wipeout
	} else {
		player.velocity = pixel.V(newVelX, player.velocity.Y)
		newX := player.position.X + player.velocity.X*scene.TimeSinceLastFrame
		newY := player.position.Y + player.velocity.Y*scene.TimeSinceLastFrame
		player.position = pixel.V(newX, newY)
	}

	camPosX := player.position.X - scene.Window.Bounds().Center().X
	camPosY := player.position.Y - scene.Window.Bounds().Center().Y
	scene.CameraPosition = pixel.V(camPosX, camPosY)
}

// updateState is where we update any game state.
// Maybe update scores, object states that aren't related to input.
func updateState(scene *Scene) {
	player := scene.Player
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
		sprite := scene.Sprites.harddrive
		if rand.Intn(2) == 0 {
			sprite = scene.Sprites.server
		}
		newObj := &Object{
			position: player.position.Add(pixel.V(float64(randX), -500)),
			velocity: pixel.V(0, 0),
			sprite:   sprite,
		}
		scene.Obstacles = append(scene.Obstacles, newObj)
		scene.TimeSinceLastObstacle = 0
	}
}

func detectCollisions(scene *Scene) {
	for _, obstacle := range scene.Obstacles {
		if intersectRect(scene.Player, obstacle) {
			scene.Dead = true
		}
	}
}

func intersectRect(object1 *Object, object2 *Object) bool {
	object1Right := object1.position.X + object1.sprite.Frame().W()
	object2Right := object2.position.X + object2.sprite.Frame().W()
	object1Bottom := object1.position.Y + object1.sprite.Frame().H()
	object2Bottom := object2.position.Y + object2.sprite.Frame().H()

	// HACK: We had to account for half of speed for some reason on detecting where
	// the player top position is. I have a feeling this has to do with
	// detecting collissions after applying speed but it's unclear.
	collides := !(object2.position.X > object1Right ||
		object2Right < object1.position.X ||
		object2.position.Y > object1Bottom ||
		object2Bottom < object1.position.Y+(speed/2))

	// Used for debug purposes.
	// fmt.Printf("[TOP: %4.0f BOTTOM: %4.0f LEFT: %4.0f RIGHT: %4.0f]\t[TOP: %4.0f BOTTOM: %4.0f LEFT: %4.0f RIGHT: %4.0f]\n",
	// 	object1.position.Y,
	// 	object1Bottom,
	// 	object1.position.X,
	// 	object1Right,
	// 	object2.position.Y,
	// 	object2Bottom,
	// 	object2.position.X,
	// 	object2Right)

	return collides
}

// render is where we render graphics after all the input and game state has been processed.
func render(scene *Scene) {
	player := scene.Player
	cameraMatrix := pixel.IM.Moved(scene.CameraPosition.Scaled(-1))
	scene.Window.SetMatrix(cameraMatrix)

	scene.Window.Clear(colornames.Blueviolet)
	player.sprite.Draw(scene.Window, pixel.IM.Moved(player.position))
	for _, o := range scene.Obstacles {
		o.sprite.Draw(scene.Window, pixel.IM.Moved(o.position))
	}

	if scene.Dead {
		// atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
		// basicTxt := text.New(pixel.V(scene.Player.position.X-200, scene.Player.position.Y+200), atlas)
		// fmt.Fprintln(basicTxt, "DEAD!!!!")
		// basicTxt.Draw(scene.Window, pixel.IM.Scaled(basicTxt.Orig, 4))
		player.sprite.Draw(scene.Window, pixel.IM.Moved(player.position))
	}
	scene.Window.Update()
}

func getSprite(img string) *pixel.Sprite {
	img = fmt.Sprintf("graphics/%s.png", img)
	playerSprite, err := graphics.LoadPicture(img)
	if err != nil {
		panic(err)
	}

	return pixel.NewSprite(playerSprite, playerSprite.Bounds())
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

	scene.Player = &Object{
		position: win.Bounds().Center(),
		velocity: pixel.V(0, -speed),
	}

	scene.Sprites = &Sprites{
		left:      getSprite("left"),
		right:     getSprite("right"),
		forward:   getSprite("forward"),
		server:    getSprite("serverrack"),
		harddrive: getSprite("harddrive"),
		jump:      getSprite("jump"),
		jumpleft:  getSprite("jumpleft"),
		jumpright: getSprite("jumpright"),
		wipeout:   getSprite("wipeout"),
	}

	scene.LastFrameTime = time.Now()
	return scene
}
