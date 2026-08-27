package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"github.com/langhuihui/gomem"
	task "github.com/langhuihui/gotask"
	"io"
	"log/slog"
	"m7s.live/v5"
	"m7s.live/v5/pkg/codec"
	"m7s.live/v5/pkg/format"
	"net"
	"os"
	"sync"
	"time"
)

var (
	errPublisherUnavailable   = errors.New("publisher unavailable")
	errAudioWriterUnavailable = errors.New("audio writer unavailable")
	errAudioFrameUnavailable  = errors.New("audio frame unavailable")
	errVideoWriterUnavailable = errors.New("video writer unavailable")
	errVideoFrameUnavailable  = errors.New("video frame unavailable")
)

type connection struct {
	conn net.Conn
	*slog.Logger
	stopChan        chan struct{}
	stopOnce        sync.Once
	publisher       *m7s.Publisher
	onJoinEvent     func(c *connection, pack *jt1078.Packet) error
	onLeaveEvent    func()
	timestampFunc   func(pack *jt1078.Packet) time.Duration
	audioWriterOnce sync.Once
	videoWriterOnce sync.Once
	audioWriter     *m7s.PublishAudioWriter[*format.Mpeg2Audio]
	videoWriter     *m7s.PublishVideoWriter[*format.AnnexB]
	debug           struct {
		hasRecord           bool
		file                *os.File
		closeTime           time.Time
		hasTemporaryStorage bool
		temporaryStorage    []byte
	}
}

func newConnection(c net.Conn, log *slog.Logger, timestampFunc func(pack *jt1078.Packet) time.Duration) *connection {
	return &connection{
		Logger:        log,
		conn:          c,
		stopChan:      make(chan struct{}),
		timestampFunc: timestampFunc,
	}
}

func (c *connection) run(ctx context.Context, waitSubscriberOverTime time.Duration) error {
	var (
		data             = make([]byte, 10*1024)
		packParse        = newPackageParse()
		once             sync.Once
		onJoinErr        error
		handleErr        error
		ticker           = time.NewTicker(time.Second)
		firstWaitSubTime time.Time
	)
	defer func() {
		packParse.clear()
		ticker.Stop()
		clear(data)
		c.stop()
		if c.debug.hasRecord {
			_ = c.debug.file.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if c.publisher != nil && waitSubscriberOverTime > 0 {
				if c.publisher.State == m7s.PublisherStateWaitSubscriber {
					if firstWaitSubTime.IsZero() {
						firstWaitSubTime = time.Now()
					} else if time.Since(firstWaitSubTime) > waitSubscriberOverTime {
						return fmt.Errorf("wait subscriber over time %s", waitSubscriberOverTime.String())
					}
				} else {
					firstWaitSubTime = time.Time{}
				}
			}
		default:
			if n, err := c.conn.Read(data); err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
					return nil
				}
				return err
			} else if n > 0 {
				if c.debug.hasTemporaryStorage {
					c.debug.temporaryStorage = append(c.debug.temporaryStorage, data[:n]...)
				}
				for pack, err := range packParse.parse(data[:n]) {
					if err == nil {
						once.Do(func() {
							onJoinErr = c.onJoinEvent(c, pack)
						})
						if onJoinErr == nil {
							if err := c.handle(pack); err != nil {
								handleErr = err
							}
						}
					} else if errors.Is(err, jt1078.ErrBodyLength2Short) || errors.Is(err, jt1078.ErrHeaderLength2Short) {
						// 数据长度不够的 忽略
					} else {
						return err
					}
				}
				if onJoinErr != nil {
					return onJoinErr
				}
				if handleErr != nil {
					return handleErr
				}
				if c.debug.hasRecord {
					if c.debug.hasTemporaryStorage {
						c.debug.hasTemporaryStorage = false
						c.saveDebugFile(c.debug.temporaryStorage)
					} else {
						c.saveDebugFile(data[:n])
					}
				}
			}
		}
	}
}

func (c *connection) saveDebugFile(data []byte) {
	if _, err := c.debug.file.WriteString(fmt.Sprintf("%x", data)); err != nil {
		c.Warn("write debug file",
			slog.String("err", err.Error()))
	}
	_ = c.debug.file.Sync()
	if time.Since(c.debug.closeTime) > 0 {
		_ = c.debug.file.Close()
		c.debug.hasRecord = false
	}
}

func (c *connection) stop() {
	c.stopOnce.Do(func() {
		close(c.stopChan)
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.onLeaveEvent()
		if c.publisher != nil {
			c.publisher.Stop(task.ErrTaskComplete)
		}
		if c.audioWriter != nil {
			c.audioWriter.Stop(task.ErrTaskComplete)
		}
		if c.videoWriter != nil {
			c.videoWriter.Stop(task.ErrTaskComplete)
		}
	})
}

func (c *connection) getAudioWriter() (*m7s.PublishAudioWriter[*format.Mpeg2Audio], error) {
	if c.publisher == nil {
		return nil, errPublisherUnavailable
	}
	c.audioWriterOnce.Do(func() {
		allocator := gomem.NewScalableMemoryAllocator(1 << gomem.MinPowerOf2)
		c.audioWriter = m7s.NewPublishAudioWriter[*format.Mpeg2Audio](c.publisher, allocator)
	})
	if c.audioWriter == nil {
		return nil, errAudioWriterUnavailable
	}
	if c.audioWriter.AudioFrame == nil {
		return nil, errAudioFrameUnavailable
	}
	return c.audioWriter, nil
}

func (c *connection) getVideoWriter() (*m7s.PublishVideoWriter[*format.AnnexB], error) {
	if c.publisher == nil {
		return nil, errPublisherUnavailable
	}
	c.videoWriterOnce.Do(func() {
		allocator := gomem.NewScalableMemoryAllocator(1 << gomem.MinPowerOf2)
		c.videoWriter = m7s.NewPublishVideoWriter[*format.AnnexB](c.publisher, allocator)
	})
	if c.videoWriter == nil {
		return nil, errVideoWriterUnavailable
	}
	if c.videoWriter.VideoFrame == nil {
		return nil, errVideoFrameUnavailable
	}
	return c.videoWriter, nil
}

func (c *connection) handle(packet *jt1078.Packet) error {
	data := packet.Body

	switch pt := packet.Flag.PT; pt {
	case jt1078.PTAAC, jt1078.PTG711A, jt1078.PTG711U:
		writer, err := c.getAudioWriter()
		if err != nil {
			return err
		}
		frame := writer.AudioFrame
		frame.ICodecCtx = &codec.AACCtx{}
		if pt == jt1078.PTG711A {
			frame.ICodecCtx = &codec.PCMACtx{}
		} else if pt == jt1078.PTG711U {
			frame.ICodecCtx = &codec.PCMUCtx{}
		}
		frame.Timestamp = c.timestampFunc(packet)
		mem := frame.NextN(len(data))
		copy(mem, data)
		return writer.NextAudio()

	case jt1078.PTH264, jt1078.PTH265:
		writer, err := c.getVideoWriter()
		if err != nil {
			return err
		}
		frame := writer.VideoFrame
		frame.Timestamp = c.timestampFunc(packet)
		// Push raw AnnexB data — do NOT use GetNalus()/PushOne() as that sets
		// HasRaw()=true and skips Demux(), which is needed to split NALUs and
		// detect SPS/PPS/IDR for GOP tracking.
		mem := frame.NextN(len(data))
		copy(mem, data)
		return writer.NextVideo()

	default:
		c.Warn("unknown pt",
			slog.Int("pt", int(pt)),
			slog.String("describe", pt.String()))
		return nil
	}
}
