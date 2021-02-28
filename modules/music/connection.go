package music

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/kkdai/youtube/v2"
)

type QueueItem struct {
	Info      *youtube.Video
	StreamURL string
}

type Connection struct {
	GuildID           string
	ChannelID         string
	EncodeOpts        *dca.EncodeOptions
	VoiceConnection   *discordgo.VoiceConnection
	Queue             []*QueueItem
	CurrentEncSession *dca.EncodeSession
	*sync.RWMutex
}

func NewConnection(voice *discordgo.VoiceConnection, opts *dca.EncodeOptions) *Connection {
	return &Connection{
		GuildID:         voice.GuildID,
		ChannelID:       voice.ChannelID,
		EncodeOpts:      opts,
		VoiceConnection: voice,
		RWMutex:         &sync.RWMutex{},
	}
}

func (c *Connection) StreamMusic() error {
	c.RLock()
	length := len(c.Queue)
	c.RUnlock()

	for i := 0; i < length; i++ {
		c.Lock()

		encodeSession, err := dca.EncodeFile(c.Queue[i].StreamURL, c.EncodeOpts)
		if err != nil {
			return err
		}

		c.CurrentEncSession = encodeSession

		c.Unlock()

		done := make(chan error)
		dca.NewStream(encodeSession, c.VoiceConnection, done)

		derr := <-done
		if derr != nil && derr != io.EOF {
			return derr
		}

		encodeSession.Cleanup()

		c.Lock()
		c.CurrentEncSession = nil
		c.Unlock()

		if len(c.Queue)-1 == i {
			for j := 0; j < 300; j++ {
				length = len(c.Queue)
				if length-1 == i {
					time.Sleep(1 * time.Second)
					continue
				}

				break
			}
		}
	}

	return nil
}

func (c *Connection) AddYouTubeVideo(url string) (*youtube.Video, error) {
	cl := youtube.Client{
		Debug: true,
	}

	video, err := cl.GetVideo(url)
	if err != nil {
		return nil, err
	}

	if len(video.Formats) == 0 {
		return nil, fmt.Errorf("no video formats found")
	}

	format := video.Formats.FindByQuality("360p")
	if format == nil {
		return nil, fmt.Errorf("no video format found")
	}

	streamUrl, err := cl.GetStreamURL(video, format)
	if err != nil {
		return nil, err
	}

	c.Lock()
	c.Queue = append(c.Queue, &QueueItem{
		Info:      video,
		StreamURL: streamUrl,
	})
	c.Unlock()

	return video, nil
}

func (c *Connection) Skip() error {
	c.RLock()
	defer c.RUnlock()

	if c.CurrentEncSession == nil {
		return nil
	}

	err := c.CurrentEncSession.Stop()
	if err != nil {
		return err
	}

	c.CurrentEncSession.Cleanup()

	return nil
}

func (c *Connection) Close(m *Music) error {
	err := c.VoiceConnection.Speaking(false)
	if err != nil {
		return err
	}

	c.VoiceConnection.Close()
	err = c.VoiceConnection.Disconnect()
	if err != nil {
		return err
	}

	delete(m.MusicConnections, c.GuildID)

	return nil
}
