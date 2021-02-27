package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/omarhachach/bear"
	"github.com/omarhachach/bear-music/modules/music"
)

func main() {
	c := make(chan os.Signal, 1)

	byt, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
		return
	}

	config := &bear.Config{}

	err = json.Unmarshal(byt, &config)
	if err != nil {
		panic(err)
		return
	}

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	b := bear.New(config).RegisterModules(&music.Music{
		MusicConnections: map[string]*music.Connection{},
	}).Start()

	<-c

	b.Close()
}
