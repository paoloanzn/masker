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
	lfoPhase   float64

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
	beatOffset := math.Mod(float64(s.sampleIndex), beatSamples)
	beatTime := beatOffset / sampleRate
	barPhase := math.Mod(float64(s.sampleIndex)/(beatSamples*4), 1.0)

	kick := softKick(beatTime)
	padL, padR := s.nextPad(barPhase)
	bedL, bedR := s.nextBed()
	textureL, textureR := s.nextTexture(rng)

	left := 0.42*padL + 0.40*float64(kick)
	right := 0.42*padR + 0.40*float64(kick)

	switch density {
	case DensityMedium:
		left += 0.24 * bedL
		right += 0.24 * bedR
	case DensityHigh:
		left += 0.24*bedL + 0.10*textureL
		right += 0.24*bedR + 0.10*textureR
	}

	leftSample := s.mixLP2L.Process(s.mixLP1L.Process(float32(left)))
	rightSample := s.mixLP2R.Process(s.mixLP1R.Process(float32(right)))

	s.sampleIndex++
	return clamp(0.70 * leftSample), clamp(0.70 * rightSample)
}

func softKick(beatTime float64) float32 {
	attack := 1.0 - math.Exp(-beatTime*160.0)
	decay := math.Exp(-beatTime * 6.2)
	fundamental := math.Sin(2.0*math.Pi*54.0*beatTime + 0.22*math.Sin(2.0*math.Pi*beatTime))
	harmonic := math.Sin(2.0 * math.Pi * 108.0 * beatTime)
	return float32(attack * decay * (0.82*fundamental + 0.18*harmonic))
}

func (s *FocusState) nextPad(barPhase float64) (float64, float64) {
	lfo := 0.94 + 0.06*math.Sin(s.lfoPhase)
	s.lfoPhase = advancePhase(s.lfoPhase, 0.045)

	frequencies := [3]float64{73.42, 110.00, 146.83}
	detune := [3]float64{0.9985, 1.0018, 0.9992}
	weights := [3]float64{0.42, 0.34, 0.24}
	swirl := 0.92 + 0.08*math.Sin(2.0*math.Pi*barPhase)

	var left, right float64
	for i := range frequencies {
		s.padPhasesL[i] = advancePhase(s.padPhasesL[i], frequencies[i]*detune[i])
		s.padPhasesR[i] = advancePhase(s.padPhasesR[i], frequencies[i]/detune[i])
		left += weights[i] * math.Sin(s.padPhasesL[i])
		right += weights[i] * math.Sin(s.padPhasesR[i])
	}

	return 0.55 * swirl * lfo * left, 0.55 * swirl * lfo * right
}

func (s *FocusState) nextBed() (float64, float64) {
	frequencies := [3]float64{174.61, 220.00, 261.63}
	detune := [3]float64{1.0009, 0.9991, 1.0012}
	weights := [3]float64{0.40, 0.35, 0.25}

	var left, right float64
	for i := range frequencies {
		s.bedPhasesL[i] = advancePhase(s.bedPhasesL[i], frequencies[i]*detune[i])
		s.bedPhasesR[i] = advancePhase(s.bedPhasesR[i], frequencies[i]/detune[i])
		left += weights[i] * math.Sin(s.bedPhasesL[i])
		right += weights[i] * math.Sin(s.bedPhasesR[i])
	}

	return 0.33 * left, 0.33 * right
}

func (s *FocusState) nextTexture(rng *xorShift32) (float64, float64) {
	const sampleRate = float64(config.SampleRate)

	subdivisionSamples := sampleRate * 60.0 / (focusTempoBPM * 2.0)
	offset := math.Mod(float64(s.sampleIndex), subdivisionSamples)
	timeInSubdivision := offset / sampleRate

	attack := 1.0 - math.Exp(-timeInSubdivision*26.0)
	decay := math.Exp(-timeInSubdivision * 5.8)
	envelope := attack * decay

	whiteL := float64(rng.nextFloat32())
	whiteR := float64(rng.nextFloat32())

	left := s.textureLP2L.Process(s.textureLP1L.Process(s.textureHP2L.Process(s.textureHP1L.Process(float32(whiteL)))))
	right := s.textureLP2R.Process(s.textureLP1R.Process(s.textureHP2R.Process(s.textureHP1R.Process(float32(whiteR)))))

	return envelope * float64(left), envelope * float64(right)
}

func advancePhase(phase, frequency float64) float64 {
	phase += 2.0 * math.Pi * frequency / config.SampleRate
	if phase >= 2.0*math.Pi {
		phase -= 2.0 * math.Pi
	}
	return phase
}
