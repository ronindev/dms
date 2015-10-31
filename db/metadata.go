package db

import (
	"github.com/anacrolix/dms/ffmpeg"
)

type Metadata struct {
	Title         string
	JpegThumbnail []byte
	PngThumbnail  []byte
	FFmpegInfo    *ffmpeg.Info
}
