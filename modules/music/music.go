package music

import (
	"github.com/omarhachach/bear"
)

type Music struct {
	MusicConnections map[string]*Connection
}

func (*Music) GetName() string {
	return "Bear Module"
}

func (*Music) GetDesc() string {
	return "Plays music from YouTube"
}

func (m *Music) GetCommands() []bear.Command {
	return []bear.Command{
		&PlayCommand{
			Music: m,
		},
		&StopCommand{
			Music: m,
		},
		&SkipCommand{
			Music: m,
		},
	}
}

func (*Music) GetVersion() string {
	return "1.0.0"
}

func (*Music) Init(*bear.Bear) {
}

func (*Music) Close(*bear.Bear) {
}
