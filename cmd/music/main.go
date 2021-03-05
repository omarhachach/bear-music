package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/omarhachach/bear"
	"github.com/omarhachach/bear-music/modules/music"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func main() {
	c := make(chan os.Signal, 1)

	byt, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
		return
	}

	config := &Config{}

	err = json.Unmarshal(byt, &config)
	if err != nil {
		panic(err)
		return
	}

	b := bear.New(config.Config)

	service, err := youtube.NewService(context.Background(), option.WithAPIKey(config.APIKey))
	if err != nil {
		b.Log.WithError(err).Error("Error creating YouTube service.")
		return
	}

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	b.RegisterModules(&music.Music{
		MusicConnections: map[string]*music.Connection{},
		Service:          service,
	}).Start()

	<-c

	b.Close()
}
