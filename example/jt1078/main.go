package main

import (
	"context"
	_ "github.com/cuteLittleDevil/m7s-jt1078/v5"
	"m7s.live/v5"
	_ "m7s.live/v5/plugin/flv"
	_ "m7s.live/v5/plugin/mp4"
	_ "m7s.live/v5/plugin/preview"
	_ "m7s.live/v5/plugin/webrtc"
)

func main() {
	_ = m7s.Run(context.Background(), "./config.yaml")
}
