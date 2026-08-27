package pkg

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"m7s.live/v5"
	"m7s.live/v5/pkg/format"
)

func newTestConnection() *connection {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	return newConnection(nil, logger, func(*jt1078.Packet) time.Duration { return 0 })
}

func TestConnectionHandleRejectsUnavailablePublisher(t *testing.T) {
	c := newTestConnection()
	err := c.handle(&jt1078.Packet{Flag: jt1078.Flag{PT: jt1078.PTH264}})
	if !errors.Is(err, errPublisherUnavailable) {
		t.Fatalf("handle() error = %v, want %v", err, errPublisherUnavailable)
	}
}

func TestConnectionHandleRejectsUnavailableWriters(t *testing.T) {
	tests := []struct {
		name string
		pt   jt1078.PTType
		want error
	}{
		{name: "audio", pt: jt1078.PTAAC, want: errAudioWriterUnavailable},
		{name: "video", pt: jt1078.PTH264, want: errVideoWriterUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestConnection()
			c.publisher = &m7s.Publisher{}
			err := c.handle(&jt1078.Packet{Flag: jt1078.Flag{PT: tt.pt}})
			if !errors.Is(err, tt.want) {
				t.Fatalf("handle() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestConnectionHandleRejectsUnavailableFrames(t *testing.T) {
	tests := []struct {
		name    string
		pt      jt1078.PTType
		prepare func(*connection)
		want    error
	}{
		{
			name: "audio",
			pt:   jt1078.PTAAC,
			prepare: func(c *connection) {
				c.publisher.PubAudio = true
				c.audioWriter = &m7s.PublishAudioWriter[*format.Mpeg2Audio]{Publisher: c.publisher}
				c.audioWriterOnce.Do(func() {})
			},
			want: errAudioFrameUnavailable,
		},
		{
			name: "video",
			pt:   jt1078.PTH264,
			prepare: func(c *connection) {
				c.publisher.PubVideo = true
				c.videoWriter = &m7s.PublishVideoWriter[*format.AnnexB]{Publisher: c.publisher}
				c.videoWriterOnce.Do(func() {})
			},
			want: errVideoFrameUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestConnection()
			c.publisher = &m7s.Publisher{}
			tt.prepare(c)
			err := c.handle(&jt1078.Packet{Flag: jt1078.Flag{PT: tt.pt}})
			if !errors.Is(err, tt.want) {
				t.Fatalf("handle() error = %v, want %v", err, tt.want)
			}
		})
	}
}

type panicReadConn struct {
	net.Conn
}

func (panicReadConn) Read([]byte) (int, error) {
	panic("test connection panic")
}

func (panicReadConn) Close() error {
	return nil
}

func TestServiceRunConnectionRecoversPanic(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	service := &Service{Logger: logger}
	client := newConnection(panicReadConn{}, logger, func(*jt1078.Packet) time.Duration { return 0 })
	client.onLeaveEvent = func() {}
	httpBody := map[string]any{"streamPath": "live/test"}

	service.runConnection(context.Background(), client, 0, &httpBody)

	select {
	case <-client.stopChan:
	default:
		t.Fatal("connection was not stopped after panic")
	}
	logOutput := output.String()
	for _, want := range []string{"connection panic", "test connection panic", "runtime/debug.Stack", "live/test"} {
		if !strings.Contains(logOutput, want) {
			t.Fatalf("panic log %q does not contain %q", logOutput, want)
		}
	}
}
