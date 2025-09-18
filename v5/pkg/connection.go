package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"io"
	"log/slog"
	"m7s.live/v5"
	"m7s.live/v5/pkg/codec"
	"m7s.live/v5/pkg/format"
	"m7s.live/v5/pkg/task"
	"m7s.live/v5/pkg/util"
	"net"
	"sync"
	"time"
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
			}
		}
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
	})
}

func (c *connection) handle(packet *jt1078.Packet) error {
	data := packet.Body

	switch pt := packet.Flag.PT; pt {
	case jt1078.PTAAC, jt1078.PTG711A, jt1078.PTG711U:
		c.audioWriterOnce.Do(func() {
			allocator := util.NewScalableMemoryAllocator(1 << util.MinPowerOf2)
			c.audioWriter = m7s.NewPublishAudioWriter[*format.Mpeg2Audio](c.publisher, allocator)
		})
		writer := c.audioWriter
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
		c.videoWriterOnce.Do(func() {
			allocator := util.NewScalableMemoryAllocator(1 << util.MinPowerOf2)
			c.videoWriter = m7s.NewPublishVideoWriter[*format.AnnexB](c.publisher, allocator)
		})
		writer := c.videoWriter
		// 为每帧创建 H26xFrame
		frame := writer.VideoFrame
		// 设置正确间隔的时间戳
		frame.Timestamp = c.timestampFunc(packet)
		// 写入 NALU 数据
		nalus := frame.GetNalus()
		// 假如 frameData 中只有一个 NALU，否则需要循环执行下面的代码
		p := nalus.GetNextPointer()
		mem := frame.NextN(len(data))
		copy(mem, data)
		p.PushOne(mem)
		return writer.NextVideo()

	default:
		c.Warn("unknown pt",
			slog.Int("pt", int(pt)),
			slog.String("describe", pt.String()))
		return nil
	}
}
