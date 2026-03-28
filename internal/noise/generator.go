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
	ModeVoice
)

func (m Mode) String() string {
	switch m {
	case ModeBrown:
		return "Brown"
	case ModePink:
		return "Pink"
	case ModeVoice:
		return "Voice-focused"
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

	voiceHPL OnePoleHP
	voiceHPR OnePoleHP
	voiceLPL OnePoleLP
	voiceLPR OnePoleLP
}

func NewGenerator() *Generator {
	generator := &Generator{
		rng: xorShift32{x: 0x12345678},

		voiceHPL: NewOnePoleHP(700),
		voiceHPR: NewOnePoleHP(700),
		voiceLPL: NewOnePoleLP(3500),
		voiceLPR: NewOnePoleLP(3500),
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
		case ModeVoice:
			left, right = g.nextVoicePair()
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

func (g *Generator) nextVoicePair() (float32, float32) {
	whiteL := g.rng.nextFloat32()
	whiteR := g.rng.nextFloat32()

	left := g.pinkL.Next(whiteL)
	right := g.pinkR.Next(whiteR)

	left = g.voiceHPL.Process(left)
	right = g.voiceHPR.Process(right)

	left = g.voiceLPL.Process(left)
	right = g.voiceLPR.Process(right)

	return clamp(1.8 * left), clamp(1.8 * right)
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
