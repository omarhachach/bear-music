package music

import (
	"github.com/omarhachach/bear"
)

type StopCommand struct {
	Music *Music
}

func (*StopCommand) GetCallers() []string {
	return []string{
		"-stop",
	}
}

func (s *StopCommand) GetHandler() func(*bear.Context) {
	return func(c *bear.Context) {
		channel, err := c.Session.Channel(c.ChannelID)
		if err != nil {
			c.Log.WithError(err).Error("Error getting channel.")
			return
		}

		conn, ok := s.Music.MusicConnections[channel.GuildID]
		if !ok {
			c.SendErrorMessage("No music stream is playing.")
			return
		}

		err = conn.Close(s.Music)
		if err != nil {
			c.Log.WithError(err).Error("Error closing music stream.")
			return
		}

		c.SendSuccessMessage("Stopped music stream.")
	}
}
