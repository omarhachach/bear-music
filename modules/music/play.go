package music

import (
	"runtime"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/omarhachach/bear"
)

type PlayCommand struct {
	Music *Music
}

func (p *PlayCommand) GetCallers() []string {
	return []string{
		"-play",
		"-p",
	}
}

func (p *PlayCommand) GetHandler() func(*bear.Context) {
	return func(c *bear.Context) {
		cmdSplit := strings.Split(c.Message.Content, " ")
		if len(cmdSplit) != 2 {
			c.SendErrorMessage("Error sending message")
			return
		}

		channel, err := c.Session.Channel(c.ChannelID)
		if err != nil {
			c.Log.WithError(err).Error("Error getting channel.")
			return
		}

		guild, err := c.Session.State.Guild(channel.GuildID)
		if err != nil {
			c.Log.WithError(err).Errorf("Error getting guild.")
			return
		}

		musicChannelID := ""
		for _, voiceState := range guild.VoiceStates {
			if voiceState.UserID == c.Message.Author.ID {
				musicChannelID = voiceState.ChannelID
				break
			}
		}

		conn, ok := p.Music.MusicConnections[channel.GuildID]
		if !ok {
			if musicChannelID == "" {
				c.SendErrorMessage("Please join a voice channel")
				return
			}

			go func() {
				voice, err := c.Session.ChannelVoiceJoin(guild.ID, musicChannelID, false, true)
				if err != nil {
					c.Log.WithError(err).Error("Error joining voice channel.")
					return
				}

				voice.LogLevel = discordgo.LogWarning

				conn := NewConnection(voice, &dca.EncodeOptions{
					Volume:         256,
					Channels:       2,
					FrameRate:      48000,
					FrameDuration:  20,
					Bitrate:        64,
					PacketLoss:     1,
					RawOutput:      true,
					Application:    dca.AudioApplicationAudio,
					CoverFormat:    "jpeg",
					BufferedFrames: 100,
					VBR:            true,
				})

				p.Music.MusicConnections[channel.GuildID] = conn

				vid, err := conn.AddYouTubeVideo(cmdSplit[1])
				if err != nil {
					c.Log.WithError(err).Debug("Error adding YouTube video to queue.")
					c.SendErrorMessage("YouTube link invalid.")
					return
				}

				for voice.Ready == false {
					runtime.Gosched()
				}

				c.SendSuccessMessage("Started playing " + vid.Title)

				err = conn.StreamMusic()
				if err != nil {
					c.Log.WithError(err).Error("Error starting music stream.")
					return
				}

				err = conn.Close(p.Music)
				if err != nil {
					c.Log.WithError(err).Error("Error closing music connection")
				}
			}()

			return
		}

		if musicChannelID != conn.ChannelID {
			c.SendErrorMessage("You are not in the voice channel with the music bot.")
			return
		}

		vid, err := conn.AddYouTubeVideo(cmdSplit[1])
		if err != nil {
			c.Log.WithError(err).Debug("Error adding YouTube video to queue.")
			c.SendErrorMessage("YouTube link invalid.")
			return
		}

		c.SendSuccessMessage("Added " + vid.Title + " to the queue.")
	}
}
