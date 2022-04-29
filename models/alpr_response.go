package models

type AlprResponseCandidate struct {
	Plate      string  `json:"plate"`
	Confidence float32 `json:"confidence"`
}

type AlprResponseCoordinate struct {
	X0 int `json:"x0"`
	Y0 int `json:"y0"`
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
}

type AlprResponseResult struct {
	Plate            string                   `json:"plate"`
	Confidence       float32                  `json:"confidence"`
	ProcessingTimeMs float32                  `json:"processing_time_ms"`
	Coordinates      AlprResponseCoordinate   `json:"coordinates"`
	Candidates       []*AlprResponseCandidate `json:"candidates"`
}

type AlprResponse struct {
	Base64Image      string                `json:"base64_image"`
	ImgWidth         int                   `json:"img_width"`
	ImgHeight        int                   `json:"img_height"`
	ProcessingTimeMs float32               `json:"processing_time_ms"`
	Results          []*AlprResponseResult `json:"results"`
	Id               string                `json:"id"`
	SourceId         string                `json:"source_id"`
	CreatedAt        string                `json:"created_at"`
	AiClipEnabled    bool                  `json:"ai_clip_enabled"`
}
