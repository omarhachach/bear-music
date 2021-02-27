package music

import (
	"github.com/omarhachach/bear"
)

type SkipCommand struct {
	Music *Music
}

func (*SkipCommand) GetCallers() []string {
	return []string{
		"-skip",
	}
}

func (s *SkipCommand) GetHandler() func(*bear.Context) {
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

		err = conn.Skip()
		if err != nil {
			c.Log.WithError(err).Error("Error skipping current song.")
			return
		}

		c.SendSuccessMessage("Skipping current song.")
	}
}

