package pkg

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func onNoticeEvent(url string, httpBody map[string]any) {
	jsonBody, err := json.Marshal(httpBody)
	if err != nil {
		slog.Warn("json marshal fail",
			slog.Any("url", url),
			slog.Any("data", httpBody),
			slog.Any("err", err))
		return
	}

	go func(body []byte) {
		client := &http.Client{
			Timeout: 3 * time.Second,
		}
		if strings.HasPrefix(strings.ToUpper(url), "HTTPS") {
			client.Transport = &http.Transport{
				DisableKeepAlives: true,
				TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
				IdleConnTimeout:   10 * time.Second, // 空闲连接超时
			}
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
		if err != nil {
			slog.Warn("req fail",
				slog.Any("url", url),
				slog.Any("data", httpBody),
				slog.Any("err", err))
			return
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		resp, err := client.Do(req)
		if err != nil {
			slog.Warn("req fail",
				slog.Any("url", url),
				slog.Any("data", httpBody),
				slog.Any("err", err))
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()
	}(jsonBody)
}
