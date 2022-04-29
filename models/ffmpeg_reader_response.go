package models

type FFmpegReaderResponse struct {
	Name          string `json:"name"`
	Source        string `json:"source"`
	Img           string `json:"img"`
	AiClipEnabled bool   `json:"ai_clip_enabled"`
}
