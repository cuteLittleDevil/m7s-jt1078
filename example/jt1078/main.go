package main

import (
	"context"
	"flag"
	_ "github.com/cuteLittleDevil/m7s-jt1078/v5"
	_ "github.com/ncruces/go-sqlite3/embed"
	"m7s.live/v5"
	_ "m7s.live/v5/plugin/flv"
	_ "m7s.live/v5/plugin/mp4"
	_ "m7s.live/v5/plugin/preview"
	_ "m7s.live/v5/plugin/webrtc"
)

func main() {
	// go run -tags sqlite main.go
	conf := flag.String("c", "config.yaml", "config file")
	flag.Parse()
	// ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second*100))
	_ = m7s.Run(context.Background(), *conf)
}
