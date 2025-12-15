package model

import "github.com/airenas/tts-line/internal/pkg/acronyms/service/api"

type Input struct {
	Word string
	MI   string
	Mode api.Mode
}
