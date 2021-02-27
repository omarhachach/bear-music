package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/omarhachach/bear"
	"github.com/omarhachach/bear-music/modules/music"
)

func main() {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	b := bear.New(&bear.Config{
		Log: &bear.LogConfig{
			Debug: true,
			File:  "",
		},
		DiscordToken: "Njc0MzYyODMwMzQ2MjU2Mzk1.XjnfUw.TE-NgSM4AU4VmTUUd0buN7AwTyE",
	}).RegisterModules(&music.Music{
		MusicConnections: map[string]*music.Connection{},
	}).Start()

	<-c

	b.Close()
}
