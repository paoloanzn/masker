package noise

import (
	"math"

	"masker/internal/config"
)

const focusTempoBPM = 72.0

type Density int32

const (
	DensityLow Density = iota
	DensityMedium
	DensityHigh
)

func (d Density) String() string {
	switch d {
	case DensityLow:
		return "Low"
	case DensityMedium:
		return "Medium"
	case DensityHigh:
		return "High"
	default:
		return "Unknown"
	}
}

func (d Density) Next() Density {
	switch d {
	case DensityLow:
		return DensityMedium
	case DensityMedium:
		return DensityHigh
	case DensityHigh:
		return DensityLow
	default:
		return DensityMedium
	}
}

func (d Density) Previous() Density {
	switch d {
	case DensityLow:
		return DensityHigh
	case DensityMedium:
		return DensityLow
	case DensityHigh:
		return DensityMedium
	default:
		return DensityMedium
	}
}

type FocusState struct {
	sampleIndex uint64

	padPhasesL [3]float64
	padPhasesR [3]float64
	bedPhasesL [3]float64
	bedPhasesR [3]float64
	padLFO     float64
	padDrift   float64
	bedDrift   float64

	textureHP1L OnePoleHP
	textureHP1R OnePoleHP
	textureHP2L OnePoleHP
	textureHP2R OnePoleHP
	textureLP1L OnePoleLP
	textureLP1R OnePoleLP
	textureLP2L OnePoleLP
	textureLP2R OnePoleLP

	mixLP1L OnePoleLP
	mixLP1R OnePoleLP
	mixLP2L OnePoleLP
	mixLP2R OnePoleLP
}

func NewFocusState() FocusState {
	return FocusState{
		textureHP1L: NewOnePoleHP(180),
		textureHP1R: NewOnePoleHP(180),
		textureHP2L: NewOnePoleHP(180),
		textureHP2R: NewOnePoleHP(180),
		textureLP1L: NewOnePoleLP(1150),
		textureLP1R: NewOnePoleLP(1150),
		textureLP2L: NewOnePoleLP(1150),
		textureLP2R: NewOnePoleLP(1150),
		mixLP1L:     NewOnePoleLP(1650),
		mixLP1R:     NewOnePoleLP(1650),
		mixLP2L:     NewOnePoleLP(1650),
		mixLP2R:     NewOnePoleLP(1650),
	}
}

func (s *FocusState) NextPair(rng *xorShift32, density Density) (float32, float32) {
	const sampleRate = float64(config.SampleRate)

	beatSamples := sampleRate * 60.0 / focusTempoBPM
	barSamples := beatSamples * 4.0
	beatOffset := math.Mod(float64(s.sampleIndex), beatSamples)
	beatTime := beatOffset / sampleRate
	barPhase := math.Mod(float64(s.sampleIndex)/barSamples, 1.0)
	barIndex := uint64(float64(s.sampleIndex) / barSamples)

	kick := softKick(beatTime)
	padL, padR := s.nextPad(barPhase)
	bedL, bedR := s.nextBed(barPhase)
	textureL, textureR := s.nextTexture(rng, barPhase, barIndex)

	left := 0.48*padL + 0.26*float64(kick)
	right := 0.48*padR + 0.26*float64(kick)

	switch density {
	case DensityMedium:
		left += 0.22 * bedL
		right += 0.22 * bedR
	case DensityHigh:
		left += 0.22*bedL + 0.055*textureL
		right += 0.22*bedR + 0.055*textureR
	}

	leftSample := s.mixLP2L.Process(s.mixLP1L.Process(float32(left)))
	rightSample := s.mixLP2R.Process(s.mixLP1R.Process(float32(right)))

	s.sampleIndex++
	return clamp(0.70 * leftSample), clamp(0.70 * rightSample)
}

func softKick(beatTime float64) float32 {
	attack := 1.0 - math.Exp(-beatTime*44.0)
	decay := math.Exp(-beatTime * 10.8)
	fundamental := math.Sin(2.0 * math.Pi * 48.0 * beatTime)
	undertone := math.Sin(2.0*math.Pi*42.0*beatTime + 0.06*math.Sin(2.0*math.Pi*0.35*beatTime))
	return float32(attack * decay * (0.78*fundamental + 0.22*undertone))
}

func (s *FocusState) nextPad(barPhase float64) (float64, float64) {
	lfo := 0.95 + 0.05*math.Sin(s.padLFO)
	s.padLFO = advancePhase(s.padLFO, 0.040)
	s.padDrift = advancePhase(s.padDrift, 0.009)

	frequencies := [3]float64{73.42, 110.00, 146.83}
	detune := [3]float64{0.9985, 1.0018, 0.9992}
	weights := [3]float64{
		0.42 + 0.018*math.Sin(s.padDrift),
		0.34 + 0.014*math.Sin(s.padDrift+2.1),
		0.24 + 0.012*math.Sin(s.padDrift+4.0),
	}
	normalizeWeights(&weights)
	swirl := 0.93 + 0.04*math.Sin(2.0*math.Pi*barPhase) + 0.02*math.Sin(s.padDrift+1.4)

	var left, right float64
	for i := range frequencies {
		s.padPhasesL[i] = advancePhase(s.padPhasesL[i], frequencies[i]*detune[i])
		s.padPhasesR[i] = advancePhase(s.padPhasesR[i], frequencies[i]/detune[i])
		left += weights[i] * math.Sin(s.padPhasesL[i])
		right += weights[i] * math.Sin(s.padPhasesR[i])
	}

	return 0.55 * swirl * lfo * left, 0.55 * swirl * lfo * right
}

func (s *FocusState) nextBed(barPhase float64) (float64, float64) {
	s.bedDrift = advancePhase(s.bedDrift, 0.006)

	frequencies := [3]float64{174.61, 220.00, 261.63}
	detune := [3]float64{1.0009, 0.9991, 1.0012}
	weights := [3]float64{
		0.40 + 0.016*math.Sin(s.bedDrift),
		0.35 + 0.014*math.Sin(s.bedDrift+2.4),
		0.25 + 0.012*math.Sin(s.bedDrift+4.2),
	}
	normalizeWeights(&weights)
	level := 0.28 * (0.92 + 0.05*math.Sin(s.bedDrift+0.8) + 0.02*math.Sin(2.0*math.Pi*barPhase))

	var left, right float64
	for i := range frequencies {
		s.bedPhasesL[i] = advancePhase(s.bedPhasesL[i], frequencies[i]*detune[i])
		s.bedPhasesR[i] = advancePhase(s.bedPhasesR[i], frequencies[i]/detune[i])
		left += weights[i] * math.Sin(s.bedPhasesL[i])
		right += weights[i] * math.Sin(s.bedPhasesR[i])
	}

	return level * left, level * right
}

func (s *FocusState) nextTexture(rng *xorShift32, barPhase float64, barIndex uint64) (float64, float64) {
	whiteL := float64(rng.nextFloat32())
	whiteR := float64(rng.nextFloat32())

	left := s.textureLP2L.Process(s.textureLP1L.Process(s.textureHP2L.Process(s.textureHP1L.Process(float32(whiteL)))))
	right := s.textureLP2R.Process(s.textureLP1R.Process(s.textureHP2R.Process(s.textureHP1R.Process(float32(whiteR)))))

	depth := textureDepth(barIndex, barPhase)

	return depth * float64(left), depth * float64(right)
}

func advancePhase(phase, frequency float64) float64 {
	phase += 2.0 * math.Pi * frequency / config.SampleRate
	if phase >= 2.0*math.Pi {
		phase -= 2.0 * math.Pi
	}
	return phase
}

func normalizeWeights(weights *[3]float64) {
	total := weights[0] + weights[1] + weights[2]
	if total == 0 {
		return
	}

	for i := range weights {
		weights[i] /= total
	}
}

func smoothstep(value float64) float64 {
	if value <= 0 {
		return 0
	}
	if value >= 1 {
		return 1
	}
	return value * value * (3.0 - 2.0*value)
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func textureDepth(barIndex uint64, barPhase float64) float64 {
	current := 0.16 + 0.14*unitHash(barIndex)
	next := 0.16 + 0.14*unitHash(barIndex+1)
	return lerp(current, next, smoothstep(barPhase))
}

func unitHash(value uint64) float64 {
	hashed := value + 0x9e3779b97f4a7c15
	hashed ^= hashed >> 30
	hashed *= 0xbf58476d1ce4e5b9
	hashed ^= hashed >> 27
	hashed *= 0x94d049bb133111eb
	hashed ^= hashed >> 31
	return float64(hashed&0xffff) / 65535.0
}
