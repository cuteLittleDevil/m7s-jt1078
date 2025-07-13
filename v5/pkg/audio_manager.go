package pkg

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"
)

type audioOperationFunc func(record map[int]*session)

type (
	AudioManager struct {
		logger            *slog.Logger
		operationFuncChan chan audioOperationFunc
		audioPorts        [2]int
		audios            map[int]*session
		// OverTime 多久没有向设备写数据就自动断开
		OverTime   time.Duration
		OnJoinURL  string
		OnLeaveURL string
	}

	session struct {
		// 是否使用
		use bool
		// 把音频数据发送给设备
		audioChan chan<- []byte
	}
)

func NewAudioManager(logger *slog.Logger, audioPorts [2]int, opts ...func(am *AudioManager)) *AudioManager {
	am := &AudioManager{
		logger:            logger,
		operationFuncChan: make(chan audioOperationFunc, 10),
		audioPorts:        audioPorts,
	}
	for _, opt := range opts {
		opt(am)
	}
	return am
}

func (am *AudioManager) Init() error {
	audios := make(map[int]*session, 10)
	for port := am.audioPorts[0]; port <= am.audioPorts[1]; port++ {
		ch := make(chan []byte, 100)
		listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			return err
		}
		go func(writeChan <-chan []byte, port int) {
			for {
				conn, err := listen.Accept()
				if err == nil {
					go func(conn net.Conn) {
						// 1. 设备连接到这个端口 发送回调
						record := map[string]any{
							"port":      port,
							"address":   conn.RemoteAddr().String(),
							"startTime": time.Now().Format(time.DateTime),
						}
						onNoticeEvent(am.OnJoinURL, record)

						// 2. 处理设备读写
						client := newDevice(conn, writeChan, am.OverTime)
						completeChan := client.run()
						if err := <-completeChan; err != nil {
							record["err"] = err.Error()
							if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
								am.logger.Debug("connection close",
									slog.Int("port", port),
									slog.Any("device addr", conn.RemoteAddr().String()),
									slog.Any("err", err))
							} else {
								am.logger.Warn("read data",
									slog.Int("port", port),
									slog.Any("device addr", conn.RemoteAddr().String()),
									slog.Any("err", err))
							}
						}

						// 3. 设备断开了
						record["endTime"] = time.Now().Format(time.DateTime)
						onNoticeEvent(am.OnLeaveURL, record)
					}(conn)
				}
			}
		}(ch, port)
		audios[port] = &session{
			use:       false,
			audioChan: ch,
		}
	}
	am.audios = audios
	return nil
}

func (am *AudioManager) Run() {
	for {
		select {
		case opFunc := <-am.operationFuncChan:
			opFunc(am.audios)
		}
	}
}

func (am *AudioManager) SendAudioData(port int, data []byte) {
	ch := make(chan struct{})
	am.operationFuncChan <- func(record map[int]*session) {
		defer close(ch)
		if v, ok := record[port]; ok {
			select {
			case v.audioChan <- data:
			default:
				am.logger.Warn("audio send fail",
					slog.Int("port", port),
					slog.String("data", fmt.Sprintf("%x", data)))
				return
			}
		}
	}
	<-ch
}
