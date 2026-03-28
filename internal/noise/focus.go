package noise

import (
	"math"

	"masker/internal/config"
)

const focusTempoBPM = 72.0

type FocusPreset int32

const (
	FocusPresetLow FocusPreset = iota
	FocusPresetMedium
	FocusPresetHigh
	FocusPresetHighCognitiveLoad
)

func (p FocusPreset) String() string {
	switch p {
	case FocusPresetLow:
		return "Low"
	case FocusPresetMedium:
		return "Medium"
	case FocusPresetHigh:
		return "High"
	case FocusPresetHighCognitiveLoad:
		return "High cognitive load"
	default:
		return "Unknown"
	}
}

func (p FocusPreset) Next() FocusPreset {
	switch p {
	case FocusPresetLow:
		return FocusPresetMedium
	case FocusPresetMedium:
		return FocusPresetHigh
	case FocusPresetHigh:
		return FocusPresetHighCognitiveLoad
	case FocusPresetHighCognitiveLoad:
		return FocusPresetLow
	default:
		return FocusPresetMedium
	}
}

func (p FocusPreset) Previous() FocusPreset {
	switch p {
	case FocusPresetLow:
		return FocusPresetHighCognitiveLoad
	case FocusPresetMedium:
		return FocusPresetLow
	case FocusPresetHigh:
		return FocusPresetMedium
	case FocusPresetHighCognitiveLoad:
		return FocusPresetHigh
	default:
		return FocusPresetMedium
	}
}

type FocusState struct {
	sampleIndex uint64

	padPhasesL   [3]float64
	padPhasesR   [3]float64
	bedPhasesL   [3]float64
	bedPhasesR   [3]float64
	pulsePhasesL [3]float64
	pulsePhasesR [3]float64
	padLFO       float64
	padDrift     float64
	bedDrift     float64
	pulseDrift   float64

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

type focusProfile struct {
	padMix     float64
	kickMix    float64
	bedMix     float64
	pulseMix   float64
	textureMix float64
	pulseDepth float64
}

func focusPresetProfile(preset FocusPreset) focusProfile {
	switch preset {
	case FocusPresetLow:
		return focusProfile{
			padMix:     0.46,
			kickMix:    0.23,
			bedMix:     0.00,
			pulseMix:   0.07,
			textureMix: 0.00,
			pulseDepth: 0.07,
		}
	case FocusPresetMedium:
		return focusProfile{
			padMix:     0.46,
			kickMix:    0.23,
			bedMix:     0.19,
			pulseMix:   0.10,
			textureMix: 0.00,
			pulseDepth: 0.09,
		}
	case FocusPresetHigh:
		return focusProfile{
			padMix:     0.45,
			kickMix:    0.22,
			bedMix:     0.24,
			pulseMix:   0.14,
			textureMix: 0.080,
			pulseDepth: 0.12,
		}
	case FocusPresetHighCognitiveLoad:
		return focusProfile{
			padMix:     0.44,
			kickMix:    0.22,
			bedMix:     0.14,
			pulseMix:   0.12,
			textureMix: 0.00,
			pulseDepth: 0.10,
		}
	default:
		return focusPresetProfile(FocusPresetMedium)
	}
}

func (s *FocusState) NextPair(rng *xorShift32, preset FocusPreset) (float32, float32) {
	const sampleRate = float64(config.SampleRate)

	beatSamples := sampleRate * 60.0 / focusTempoBPM
	barSamples := beatSamples * 4.0
	beatOffset := math.Mod(float64(s.sampleIndex), beatSamples)
	beatTime := beatOffset / sampleRate
	beatPhase := beatOffset / beatSamples
	barPhase := math.Mod(float64(s.sampleIndex)/barSamples, 1.0)
	barIndex := uint64(float64(s.sampleIndex) / barSamples)
	profile := focusPresetProfile(preset)

	kick := softKick(beatTime)
	padL, padR := s.nextPad(barPhase, preset)
	bedL, bedR := s.nextBed(barPhase, preset)
	pulseL, pulseR := s.nextPulseCarrier(preset)
	textureL, textureR := s.nextTexture(rng, barPhase, barIndex)
	pulseContour := structuredPulseContour(beatPhase, profile.pulseDepth)

	left := profile.padMix*padL + profile.kickMix*float64(kick)
	right := profile.padMix*padR + profile.kickMix*float64(kick)

	left += profile.bedMix*bedL + profile.pulseMix*pulseContour*pulseL + profile.textureMix*textureL
	right += profile.bedMix*bedR + profile.pulseMix*pulseContour*pulseR + profile.textureMix*textureR

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

func structuredPulseContour(beatPhase, depth float64) float64 {
	if depth <= 0 {
		return 1.0
	}

	const attackPhase = 0.18
	if beatPhase <= attackPhase {
		return 1.0 + depth*smoothstep(beatPhase/attackPhase)
	}

	releasePhase := (beatPhase - attackPhase) / (1.0 - attackPhase)
	return 1.0 + depth*math.Exp(-5.2*releasePhase)
}

func (s *FocusState) nextPad(barPhase float64, preset FocusPreset) (float64, float64) {
	lfo := 0.95 + 0.05*math.Sin(s.padLFO)
	padDriftRate := 0.009
	swirl := 0.93 + 0.04*math.Sin(2.0*math.Pi*barPhase) + 0.02*math.Sin(s.padDrift+1.4)
	level := 0.55
	frequencies := [3]float64{73.42, 110.00, 146.83}
	detune := [3]float64{0.9985, 1.0018, 0.9992}
	weights := [3]float64{
		0.42 + 0.018*math.Sin(s.padDrift),
		0.34 + 0.014*math.Sin(s.padDrift+2.1),
		0.24 + 0.012*math.Sin(s.padDrift+4.0),
	}

	if preset == FocusPresetHighCognitiveLoad {
		lfo = 0.975 + 0.015*math.Sin(s.padLFO)
		padDriftRate = 0.003
		swirl = 0.96 + 0.01*math.Sin(s.padDrift+0.8)
		level = 0.50
		detune = [3]float64{0.9992, 1.0008, 0.9998}
		weights = [3]float64{0.56, 0.28, 0.16}
	}

	s.padLFO = advancePhase(s.padLFO, 0.040)
	s.padDrift = advancePhase(s.padDrift, padDriftRate)
	normalizeWeights(&weights)

	var left, right float64
	for i := range frequencies {
		s.padPhasesL[i] = advancePhase(s.padPhasesL[i], frequencies[i]*detune[i])
		s.padPhasesR[i] = advancePhase(s.padPhasesR[i], frequencies[i]/detune[i])
		left += weights[i] * math.Sin(s.padPhasesL[i])
		right += weights[i] * math.Sin(s.padPhasesR[i])
	}

	return level * swirl * lfo * left, level * swirl * lfo * right
}

func (s *FocusState) nextBed(barPhase float64, preset FocusPreset) (float64, float64) {
	bedDriftRate := 0.006
	frequencies := [3]float64{174.61, 220.00, 261.63}
	detune := [3]float64{1.0009, 0.9991, 1.0012}
	weights := [3]float64{
		0.40 + 0.016*math.Sin(s.bedDrift),
		0.35 + 0.014*math.Sin(s.bedDrift+2.4),
		0.25 + 0.012*math.Sin(s.bedDrift+4.2),
	}
	level := 0.28 * (0.92 + 0.05*math.Sin(s.bedDrift+0.8) + 0.02*math.Sin(2.0*math.Pi*barPhase))

	if preset == FocusPresetHighCognitiveLoad {
		bedDriftRate = 0.002
		frequencies = [3]float64{146.83, 220.00, 293.66}
		detune = [3]float64{1.0003, 0.9997, 1.0004}
		weights = [3]float64{0.55, 0.30, 0.15}
		level = 0.22 * (0.96 + 0.015*math.Sin(s.bedDrift+0.5))
	}

	s.bedDrift = advancePhase(s.bedDrift, bedDriftRate)
	normalizeWeights(&weights)

	var left, right float64
	for i := range frequencies {
		s.bedPhasesL[i] = advancePhase(s.bedPhasesL[i], frequencies[i]*detune[i])
		s.bedPhasesR[i] = advancePhase(s.bedPhasesR[i], frequencies[i]/detune[i])
		left += weights[i] * math.Sin(s.bedPhasesL[i])
		right += weights[i] * math.Sin(s.bedPhasesR[i])
	}

	return level * left, level * right
}

func (s *FocusState) nextPulseCarrier(preset FocusPreset) (float64, float64) {
	pulseDriftRate := 0.004
	frequencies := [3]float64{98.00, 123.47, 146.83}
	detune := [3]float64{0.9994, 1.0007, 0.9998}
	weights := [3]float64{
		0.48 + 0.008*math.Sin(s.pulseDrift),
		0.31 + 0.006*math.Sin(s.pulseDrift+2.2),
		0.21 + 0.005*math.Sin(s.pulseDrift+4.1),
	}
	level := 0.18

	if preset == FocusPresetHigh {
		level = 0.19
	}
	if preset == FocusPresetHighCognitiveLoad {
		pulseDriftRate = 0.0015
		frequencies = [3]float64{73.42, 110.00, 146.83}
		detune = [3]float64{0.9998, 1.0002, 1.0001}
		weights = [3]float64{0.60, 0.27, 0.13}
		level = 0.16
	}

	s.pulseDrift = advancePhase(s.pulseDrift, pulseDriftRate)
	normalizeWeights(&weights)

	var left, right float64
	for i := range frequencies {
		s.pulsePhasesL[i] = advancePhase(s.pulsePhasesL[i], frequencies[i]*detune[i])
		s.pulsePhasesR[i] = advancePhase(s.pulsePhasesR[i], frequencies[i]/detune[i])
		left += weights[i] * math.Sin(s.pulsePhasesL[i])
		right += weights[i] * math.Sin(s.pulsePhasesR[i])
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
