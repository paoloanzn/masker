package noise

import (
	"math"
	"sync/atomic"

	"masker/internal/config"
)

type Mode int32

const (
	ModeBrown Mode = iota
	ModePink
	ModeSpeech
)

func (m Mode) String() string {
	switch m {
	case ModeBrown:
		return "Brown"
	case ModePink:
		return "Pink"
	case ModeSpeech:
		return "Speech-shaped"
	default:
		return "Unknown"
	}
}

type Generator struct {
	rng xorShift32

	modeBits   atomic.Int32
	volumeBits atomic.Uint32

	brownL float32
	brownR float32

	pinkL PinkState
	pinkR PinkState

	speechL SpeechShaper
	speechR SpeechShaper
}

func NewGenerator() *Generator {
	generator := &Generator{
		rng: xorShift32{x: 0x12345678},

		speechL: NewSpeechShaper(),
		speechR: NewSpeechShaper(),
	}
	generator.SetMode(ModeBrown)
	generator.SetVolume(config.DefaultVolume)
	return generator
}

func (g *Generator) Mode() Mode {
	return Mode(g.modeBits.Load())
}

func (g *Generator) SetMode(mode Mode) {
	g.modeBits.Store(int32(mode))
}

func (g *Generator) Volume() float32 {
	return math.Float32frombits(g.volumeBits.Load())
}

func (g *Generator) SetVolume(volume float32) {
	if volume < config.MinVolume {
		volume = config.MinVolume
	}
	if volume > config.MaxVolume {
		volume = config.MaxVolume
	}
	g.volumeBits.Store(math.Float32bits(volume))
}

func (g *Generator) Fill(samples []float32) {
	volume := g.Volume()
	mode := g.Mode()

	for i := 0; i < len(samples); i += 2 {
		var left, right float32
		switch mode {
		case ModeBrown:
			left, right = g.nextBrownPair()
		case ModePink:
			left, right = g.nextPinkPair()
		case ModeSpeech:
			left, right = g.nextSpeechPair()
		default:
			left, right = g.nextBrownPair()
		}

		samples[i] = volume * left
		samples[i+1] = volume * right
	}
}

func (g *Generator) nextBrownPair() (float32, float32) {
	whiteL := g.rng.nextFloat32()
	whiteR := g.rng.nextFloat32()

	g.brownL = 0.995*g.brownL + 0.02*whiteL
	g.brownR = 0.995*g.brownR + 0.02*whiteR

	return clamp(g.brownL), clamp(g.brownR)
}

func (g *Generator) nextPinkPair() (float32, float32) {
	whiteL := g.rng.nextFloat32()
	whiteR := g.rng.nextFloat32()
	return clamp(g.pinkL.Next(whiteL)), clamp(g.pinkR.Next(whiteR))
}

func (g *Generator) nextSpeechPair() (float32, float32) {
	whiteL := g.rng.nextFloat32()
	whiteR := g.rng.nextFloat32()

	left := g.speechL.Process(whiteL)
	right := g.speechR.Process(whiteR)

	return clamp(1.35 * left), clamp(1.35 * right)
}

type xorShift32 struct {
	x uint32
}

func (r *xorShift32) nextFloat32() float32 {
	x := r.x
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	r.x = x

	uniform := float32(x) / float32(math.MaxUint32)
	return 2*uniform - 1
}

type PinkState struct {
	b0 float32
	b1 float32
	b2 float32
	b3 float32
	b4 float32
	b5 float32
	b6 float32
}

func (p *PinkState) Next(white float32) float32 {
	p.b0 = 0.99886*p.b0 + 0.0555179*white
	p.b1 = 0.99332*p.b1 + 0.0750759*white
	p.b2 = 0.96900*p.b2 + 0.1538520*white
	p.b3 = 0.86650*p.b3 + 0.3104856*white
	p.b4 = 0.55000*p.b4 + 0.5329522*white
	p.b5 = -0.7616*p.b5 - 0.0168980*white
	pink := p.b0 + p.b1 + p.b2 + p.b3 + p.b4 + p.b5 + p.b6 + 0.5362*white
	p.b6 = 0.115926 * white
	return 0.11 * pink
}

type SpeechBand struct {
	gain  float32
	hasHP bool
	hp    OnePoleHP
	lp    OnePoleLP
}

func NewSpeechBand(lowCutHz, highCutHz, gain float32) SpeechBand {
	band := SpeechBand{
		gain: gain,
		lp:   NewOnePoleLP(highCutHz),
	}
	if lowCutHz > 0 {
		band.hasHP = true
		band.hp = NewOnePoleHP(lowCutHz)
	}
	return band
}

func (b *SpeechBand) Process(sample float32) float32 {
	if b.hasHP {
		sample = b.hp.Process(sample)
	}
	return b.gain * b.lp.Process(sample)
}

type SpeechShaper struct {
	bands [6]SpeechBand
}

func NewSpeechShaper() SpeechShaper {
	// Approximate a long-term average speech spectrum with broad octave bands.
	return SpeechShaper{
		bands: [6]SpeechBand{
			NewSpeechBand(125, 250, 0.18),
			NewSpeechBand(250, 500, 0.40),
			NewSpeechBand(500, 1000, 0.85),
			NewSpeechBand(1000, 2000, 1.00),
			NewSpeechBand(2000, 4000, 0.63),
			NewSpeechBand(4000, 8000, 0.28),
		},
	}
}

func (s *SpeechShaper) Process(sample float32) float32 {
	var shaped float32
	for i := range s.bands {
		shaped += s.bands[i].Process(sample)
	}
	return shaped
}

type OnePoleHP struct {
	alpha float32
	xPrev float32
	yPrev float32
}

func NewOnePoleHP(cutoffHz float32) OnePoleHP {
	rc := 1.0 / (2.0 * math.Pi * float64(cutoffHz))
	dt := 1.0 / config.SampleRate
	alpha := float32(rc / (rc + dt))
	return OnePoleHP{alpha: alpha}
}

func (f *OnePoleHP) Process(sample float32) float32 {
	output := f.alpha * (f.yPrev + sample - f.xPrev)
	f.xPrev = sample
	f.yPrev = output
	return output
}

type OnePoleLP struct {
	alpha float32
	y     float32
}

func NewOnePoleLP(cutoffHz float32) OnePoleLP {
	dt := 1.0 / config.SampleRate
	rc := 1.0 / (2.0 * math.Pi * float64(cutoffHz))
	alpha := float32(dt / (rc + dt))
	return OnePoleLP{alpha: alpha}
}

func (f *OnePoleLP) Process(sample float32) float32 {
	f.y += f.alpha * (sample - f.y)
	return f.y
}

func clamp(sample float32) float32 {
	if sample > 1 {
		return 1
	}
	if sample < -1 {
		return -1
	}
	return sample
}
