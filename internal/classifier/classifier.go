package classifier

import (
	"context"

	"tokenlaunch/internal/domain"
)

type Classification string

const (
	ClassificationLaunch      Classification = "launch"
	ClassificationEndorsement Classification = "endorsement"
	ClassificationNone        Classification = "none"
)

type Result struct {
	Classification Classification
	Token          string
	Confidence     float64
	Reason         string
}

type Classifier interface {
	Classify(ctx context.Context, msg domain.Message) (*Result, error)
}
