package pkg

import (
	"bytes"
	"golang.org/x/exp/slog"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var (
	httpServer  *httptest.Server
	httpsServer *httptest.Server
	logOutput   bytes.Buffer
)

func TestMain(m *testing.M) {
	logger := slog.New(slog.NewJSONHandler(&logOutput, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				return a
			}
			return slog.Attr{}
		},
	}))
	slog.SetDefault(logger)

	handle := func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		if _, err = w.Write(body); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		slog.Debug(string(body))
	}
	// Create router
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/jt808/event/join-audio", handle)
	mux.HandleFunc("/api/v1/jt808/event/leave-audio", handle)
	mux.HandleFunc("/api/v1/jt808/event/real-time-join", handle)
	mux.HandleFunc("/api/v1/jt808/event/real-time-leave", handle)
	mux.HandleFunc("/api/v1/play-back-join", handle)
	mux.HandleFunc("/api/v1/play-back-leave", handle)

	httpServer = httptest.NewServer(mux)
	defer httpServer.Close()

	httpsServer = httptest.NewTLSServer(mux)
	defer httpsServer.Close()

	m.Run()
}

func Test_onNoticeEvent(t *testing.T) {
	type args struct {
		url      string
		httpBody map[string]any
	}
	type want struct {
		result string
	}
	tests := []struct {
		name string
		args
		want
	}{
		{
			name: "https 设备连接到音频端口的回调",
			args: args{
				url: httpsServer.URL + "/api/v1/jt808/event/join-audio",
				httpBody: map[string]any{
					"port":      12345,
					"address":   "192.168.1.2:8080",
					"startTime": "2025-07-13 18:15:56",
				},
			},
			want: want{
				result: `{"msg":"{"address":"192.168.1.2:8080","port":12345,"startTime":"2025-07-13 18:15:56"}"}`,
			},
		},
		{
			name: "https 设备断开了音频端口的回调",
			args: args{
				url: httpsServer.URL + "/api/v1/jt808/event/leave-audio",
				httpBody: map[string]any{
					"port":      12345,
					"address":   "192.168.1.2:8080",
					"startTime": "2025-07-13 18:15:56",
					"err":       "",
					"endTime":   "2025-07-13 19:15:56",
				},
			},
			want: want{
				result: `{"msg":"{"address":"192.168.1.2:8080","endTime":"2025-07-13 19:15:56","err":"","port":12345,"startTime":"2025-07-13 18:15:56"}"}`,
			},
		},
		{
			name: "http 设备连接到了实时视频指定端口的回调",
			args: args{
				url: httpServer.URL + "/api/v1/jt808/event/real-time-join",
				httpBody: map[string]any{
					"streamPath": "live/jt1078-295696659617-1",
					"sim":        "295696659617",
					"channel":    1,
					"startTime":  "2025-07-13 18:58:15",
				},
			},
			want: want{
				result: `{"msg":"{"channel":1,"sim":"295696659617","startTime":"2025-07-13 18:58:15","streamPath":"live/jt1078-295696659617-1"}"}`,
			},
		},
		{
			name: "http 设备断开了实时视频指定端口的回调",
			args: args{
				url: httpServer.URL + "/api/v1/jt808/event/real-time-leave",
				httpBody: map[string]any{
					"streamPath": "live/jt1078-295696659617-1",
					"sim":        "295696659617",
					"channel":    1,
					"startTime":  "2025-07-13 18:58:15",
					"endTime":    "2025-07-13 18:59:37",
				},
			},
			want: want{
				result: `{"msg":"{"channel":1,"endTime":"2025-07-13 18:59:37","sim":"295696659617","startTime":"2025-07-13 18:58:15","streamPath":"live/jt1078-295696659617-1"}"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logOutput.Reset()
			onNoticeEvent(tt.url, tt.httpBody)
			time.Sleep(time.Second)
			got := logOutput.String()
			got = strings.ReplaceAll(got, "\\", "")
			got = strings.ReplaceAll(got, "\n", "")
			if got != tt.want.result {
				t.Errorf("onNoticeEvent()\n got=%v \n want %v", got, tt.want.result)
				return
			}
		})
	}
}
