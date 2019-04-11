package main

import (
	"image"
	_ "image/png"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

const (
	windowWidth  = 1024
	windowHeight = 768
	speed        = 30
)

// Object represents an item in the game (player, obstacle, etc...)
type Object struct {
	position pixel.Vec
	velocity pixel.Vec
	sprite   *pixel.Sprite
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "SNOboard",
		Bounds: pixel.R(0, 0, windowWidth, windowHeight),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	pic, err := loadPicture("gopher-happy.png")
	if err != nil {
		panic(err)
	}

	player := &Object{
		position: win.Bounds().Center(),
		velocity: pixel.V(0, -speed),
		sprite:   pixel.NewSprite(pic, pic.Bounds()),
	}

	temp := &Object{
		position: win.Bounds().Center(),
		velocity: pixel.V(0, 0),
		sprite:   pixel.NewSprite(pic, pic.Bounds()),
	}
	// win.Clear(colornames.Skyblue)
	// player.sprite.Draw(win, pixel.IM.Moved(player.position))

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()
		cam := pixel.IM.Moved(win.Bounds().Center().Sub(player.position))
		win.SetMatrix(cam)

		newVelX := float64(0)
		if win.Pressed(pixelgl.KeyLeft) {
			newVelX -= speed
		}
		if win.Pressed(pixelgl.KeyRight) {
			newVelX += speed
		}

		player.velocity = pixel.V(newVelX, player.velocity.Y)

		newX := player.position.X + player.velocity.X*dt
		newY := player.position.Y + player.velocity.Y*dt
		player.position = pixel.V(newX, newY)

		win.Clear(colornames.Blueviolet)
		temp.sprite.Draw(win, pixel.IM.Moved(temp.position))
		player.sprite.Draw(win, pixel.IM.Moved(player.position))

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
