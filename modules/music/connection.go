package music

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/kkdai/youtube/v2"
	"github.com/sheerun/queue"
	yt "google.golang.org/api/youtube/v3"
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
	Queue             *queue.Queue
	CurrentEncSession *dca.EncodeSession
	*sync.Mutex
}

func NewConnection(voice *discordgo.VoiceConnection, opts *dca.EncodeOptions) *Connection {
	return &Connection{
		GuildID:         voice.GuildID,
		ChannelID:       voice.ChannelID,
		EncodeOpts:      opts,
		VoiceConnection: voice,
		Queue:           queue.New(),
		Mutex:           &sync.Mutex{},
	}
}

func (c *Connection) StreamMusic() error {
	for {
		item, ok := c.Queue.Pop().(*QueueItem)
		if !ok {
			return fmt.Errorf("error casting into queue item")
		}

		encodeSession, err := dca.EncodeFile(item.StreamURL, c.EncodeOpts)
		if err != nil {
			return err
		}

		c.Lock()
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

		if c.Queue.Length() == 0 {
			for i := 0; i < 150; i++ {
				time.Sleep(2 * time.Second)

				if c.Queue.Length() > 0 {
					break
				}
			}
		}

		if c.Queue.Length() == 0 {
			break
		}
	}

	return nil
}

func (c *Connection) AddYouTubeVideo(search string, service *yt.Service) (*youtube.Video, error) {
	slcall := service.Search.List([]string{"snippet"})

	res, err := slcall.Q(search).Do()
	if err != nil {
		return nil, err
	}

	cl := youtube.Client{
		Debug: true,
	}

	video, err := cl.GetVideo(res.Items[0].Id.VideoId)
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

	c.Queue.Append(&QueueItem{
		Info:      video,
		StreamURL: streamUrl,
	})

	return video, nil
}

func (c *Connection) Skip() error {
	c.Lock()
	defer c.Unlock()

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

	c.Queue.Clean()

	return nil
}
