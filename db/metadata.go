package db

import (
	"github.com/anacrolix/dms/ffmpeg"
)

type Metadata struct {
	Title      string
	FFmpegInfo *ffmpeg.Info
}
