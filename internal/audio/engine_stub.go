//go:build !darwin

package audio

import "fmt"

type Engine struct {
	filler SampleFiller
}

func NewEngine(filler SampleFiller) *Engine {
	return &Engine{filler: filler}
}

func (e *Engine) Start() error {
	return fmt.Errorf("audio engine is only supported on darwin")
}

func (e *Engine) Stop() {}
