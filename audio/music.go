package audio

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Music struct {
	deadStreamer beep.StreamSeekCloser
}

func NewMusic() *Music {
	f, err := os.Open("./dead.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, _, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	//defer streamer.Close()

	ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   0.0,
		Silent:   false,
	}
	speaker.Play(volume)

	return &Music{
		deadStreamer: streamer,
	}
}

func (music Music) PlayBackgroundMusic() {
	f, err := os.Open("./danger_zone.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   +0.0,
		Silent:   false,
	}
	speaker.Play(volume)

	_ = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	//speaker.Play(streamer)

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	<-done
}

func (music Music) PlayDeadSound() {
	_ = music.deadStreamer.Seek(100500.0)

	done := make(chan bool)
	speaker.Play(beep.Seq(music.deadStreamer, beep.Callback(func() {
		done <- true
	})))
	<-done
}
