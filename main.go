package main

import (
	"fmt"
	_ "image/png"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"storj.io/snoboard/graphics"
)

const (
	windowWidth  = 1024
	windowHeight = 768
	speed        = 300
	jumpSpeed    = 500
	jumpTime     = 0.9
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
	Difficulty            float64
	Sprites               *Sprites
	Dead                  bool
	Jumping               bool
	TimeSinceJump         float64
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

// Scoreboard represents the object rendering the player's score and other related information
type Scoreboard struct {
	score string
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

func updateScore(sc *Scene) {
	distance := sc.Player.position.Y
	score := distance * -1 / 2

	if distance > 0 {
		score = 0
	}

	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	basicTxt := text.New(sc.CameraPosition.Add(pixel.V(750, 725)), basicAtlas)
	basicTxt.Color = colornames.Black
	fmt.Fprintf(basicTxt, "Score: %s\n", strconv.FormatFloat(score, 'f', 0, 64))

	basicTxt.Draw(sc.Window, pixel.IM.Scaled(basicTxt.Orig, 2))
}

// processInput is where we process any input events from the keyboard.
func processInput(scene *Scene) {
	newVelX := float64(0)
	if scene.Window.Pressed(pixelgl.KeyLeft) {
		newVelX -= speed
		scene.Player.sprite = scene.Sprites.left
		if scene.Jumping {
			scene.Player.sprite = scene.Sprites.jumpleft
		}
	}
	if scene.Window.Pressed(pixelgl.KeyRight) {
		newVelX += speed
		scene.Player.sprite = scene.Sprites.right
		if scene.Jumping {
			scene.Player.sprite = scene.Sprites.jumpright
		}
	}
	if !scene.Window.Pressed(pixelgl.KeyRight) && !scene.Window.Pressed(pixelgl.KeyLeft) {
		scene.Player.sprite = scene.Sprites.forward
		if scene.Jumping {
			scene.Player.sprite = scene.Sprites.jump
		}
	}
	if scene.Window.Pressed(pixelgl.KeySpace) && !scene.Jumping {
		scene.Jumping = true
		scene.TimeSinceJump = 0
		scene.Player.velocity = pixel.V(0, -jumpSpeed)
	}
	if scene.Window.Pressed(pixelgl.KeyEnter) && scene.Dead {
		scene.Dead = false
		scene.Player.position = scene.Window.Bounds().Center()
		scene.Jumping = false
		scene.Difficulty = 1
		scene.Obstacles = []*Object{}
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
	scene.CameraPosition = pixel.V(camPosX, camPosY-200)
}

// updateState is where we update any game state.
// Maybe update scores, object states that aren't related to input.
func updateState(scene *Scene) {
	if scene.Dead {
		return
	}
	if scene.Jumping {
		scene.TimeSinceJump += scene.TimeSinceLastFrame
		if scene.TimeSinceJump > jumpTime {
			scene.Jumping = false
			scene.Player.velocity = pixel.V(0, -speed)
		}
	}
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
	if scene.TimeSinceLastObstacle > scene.Difficulty {
		randX := rand.Intn(2*windowWidth) - windowWidth
		sprite := scene.Sprites.harddrive
		if rand.Intn(2) == 0 {
			sprite = scene.Sprites.server
		}
		newObj := &Object{
			position: player.position.Add(pixel.V(float64(randX), -700)),
			velocity: pixel.V(0, 0),
			sprite:   sprite,
		}
		scene.Obstacles = append(scene.Obstacles, newObj)
		scene.TimeSinceLastObstacle = 0
		scene.Difficulty -= 0.01
		if scene.Difficulty < 0.25 {
			scene.Difficulty = 0.25
		}
	}
}

func detectCollisions(scene *Scene) {
	for _, obstacle := range scene.Obstacles {
		if obstacle.sprite == scene.Sprites.harddrive && scene.Jumping {
			continue
		}
		if intersectRect(scene.Player, obstacle) {
			scene.Dead = true
		}
	}
}

func intersectRect(object1 *Object, object2 *Object) bool {
	// object1Right := object1.position.X + object1.sprite.Frame().W()
	// object2Right := object2.position.X + object2.sprite.Frame().W()
	// object1Bottom := object1.position.Y + object1.sprite.Frame().H()
	// object2Bottom := object2.position.Y + object2.sprite.Frame().H()

	// // HACK: We had to account for half of speed for some reason on detecting where
	// // the player top position is. I have a feeling this has to do with
	// // detecting collissions after applying speed but it's unclear.
	// collides := !(object2.position.X > object1Right ||
	// 	object2Right < object1.position.X ||
	// 	object2.position.Y > object1Bottom ||
	// 	object2Bottom < object1.position.Y+(speed/2))

	// // Used for debug purposes.
	// // fmt.Printf("[TOP: %4.0f BOTTOM: %4.0f LEFT: %4.0f RIGHT: %4.0f]\t[TOP: %4.0f BOTTOM: %4.0f LEFT: %4.0f RIGHT: %4.0f]\n",
	// // 	object1.position.Y,
	// // 	object1Bottom,
	// // 	object1.position.X,
	// // 	object1Right,
	// // 	object2.position.Y,
	// // 	object2Bottom,
	// // 	object2.position.X,
	// // 	object2Right)

	// return collides

	minXOffset := object1.sprite.Frame().W()/2 + object2.sprite.Frame().W()/2
	minYOffset := object1.sprite.Frame().H()/2 + object2.sprite.Frame().H()/2

	xOffset := (object1.position.X + object1.sprite.Frame().W()/2) - (object2.position.X + object2.sprite.Frame().W()/2)
	yOffset := (object1.position.Y - object1.sprite.Frame().H()/2) - (object2.position.Y - object2.sprite.Frame().H()/2)

	xDiff := minXOffset - math.Abs(xOffset)
	yDiff := minYOffset - math.Abs(yOffset)

	if xDiff >= 0 && yDiff >= 0 {
		return true
	}
	return false
}

// render is where we render graphics after all the input and game state has been processed.
func render(scene *Scene) {
	player := scene.Player
	cameraMatrix := pixel.IM.Moved(scene.CameraPosition.Scaled(-1))
	scene.Window.SetMatrix(cameraMatrix)

	scene.Window.Clear(colornames.Blueviolet)
	for _, o := range scene.Obstacles {
		o.sprite.Draw(scene.Window, pixel.IM.Moved(o.position))
	}
	player.sprite.Draw(scene.Window, pixel.IM.Moved(player.position))

	if scene.Dead {
		atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
		basicTxt := text.New(pixel.V(scene.Player.position.X, scene.Player.position.Y), atlas)
		fmt.Fprintln(basicTxt, "DEAD!!!!")
		player.sprite.Draw(scene.Window, pixel.IM.Moved(player.position))
		basicTxt.Draw(scene.Window, pixel.IM.Scaled(basicTxt.Orig, 4))
	}
	updateScore(scene)
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

	scene.Difficulty = 1
	return scene
}
