package data

import (
	"encoding/json"
	"fmt"
	"io"
)

func SaveLevel(out io.Writer, level Level) error {
	if err := json.NewEncoder(out).Encode(level); err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}
	return nil
}

func LoadLevel(in io.Reader) (Level, error) {
	var level Level
	if err := json.NewDecoder(in).Decode(&level); err != nil {
		return Level{}, fmt.Errorf("failed to decode json: %w", err)
	}
	return level, nil
}

type Level struct {
	Textures []string `json:"textures"`
	Walls    []Wall   `json:"walls"`
}

type Wall struct {
	LeftEdgeX  float32 `json:"lx"`
	LeftEdgeZ  float32 `json:"lz"`
	RightEdgeX float32 `json:"rx"`
	RightEdgeZ float32 `json:"rz"`

	Ceiling *Extrusion `json:"c,omitempty"`
	Floor   *Extrusion `json:"f,omitempty"`

	FrontWall int `json:"fw"`
	BackWall  int `json:"bw"`
}

type Extrusion struct {
	Top    float32 `json:"t"`
	Bottom float32 `json:"b"`

	OuterTexture int `json:"ot"`
	FaceTexture  int `json:"ft"`
	InnerTexture int `json:"it"`
}
