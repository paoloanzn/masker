package noise

import (
	"math"
	"testing"

	"masker/internal/config"
)

func TestModeString(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{mode: ModeFocus, want: "Focus"},
		{mode: ModeADHD, want: "ADHD / Attention"},
		{mode: ModeBrown, want: "Brown"},
		{mode: ModePink, want: "Pink"},
		{mode: ModeSpeech, want: "Speech-shaped"},
		{mode: Mode(99), want: "Unknown"},
	}

	for _, test := range tests {
		if got := test.mode.String(); got != test.want {
			t.Fatalf("mode %v string = %q, want %q", test.mode, got, test.want)
		}
	}
}

func TestModeCycle(t *testing.T) {
	tests := []struct {
		name     string
		start    Mode
		next     Mode
		previous Mode
	}{
		{name: "focus wraps", start: ModeFocus, next: ModeADHD, previous: ModeSpeech},
		{name: "adhd wraps", start: ModeADHD, next: ModeBrown, previous: ModeFocus},
		{name: "brown wraps", start: ModeBrown, next: ModePink, previous: ModeADHD},
		{name: "pink wraps", start: ModePink, next: ModeSpeech, previous: ModeBrown},
		{name: "speech wraps", start: ModeSpeech, next: ModeFocus, previous: ModePink},
		{name: "unknown falls back", start: Mode(99), next: ModeFocus, previous: ModeFocus},
	}

	for _, test := range tests {
		if got := test.start.Next(); got != test.next {
			t.Fatalf("%s next = %v, want %v", test.name, got, test.next)
		}
		if got := test.start.Previous(); got != test.previous {
			t.Fatalf("%s previous = %v, want %v", test.name, got, test.previous)
		}
	}
}

func TestSetVolumeClampsToConfiguredRange(t *testing.T) {
	generator := NewGenerator()

	generator.SetVolume(config.MaxVolume + 1)
	if got := generator.Volume(); got != config.MaxVolume {
		t.Fatalf("volume = %f, want %f", got, config.MaxVolume)
	}

	generator.SetVolume(config.MinVolume - 1)
	if got := generator.Volume(); got != config.MinVolume {
		t.Fatalf("volume = %f, want %f", got, config.MinVolume)
	}
}

func TestDefaultFocusModeAndDensity(t *testing.T) {
	generator := NewGenerator()

	if got := generator.Mode(); got != ModeFocus {
		t.Fatalf("default mode = %v, want %v", got, ModeFocus)
	}
	if got := generator.FocusPreset(); got != FocusPresetMedium {
		t.Fatalf("default focus preset = %v, want %v", got, FocusPresetMedium)
	}
}

func TestFocusPresetCycle(t *testing.T) {
	tests := []struct {
		name     string
		start    FocusPreset
		next     FocusPreset
		previous FocusPreset
		want     string
	}{
		{name: "low wraps", start: FocusPresetLow, next: FocusPresetMedium, previous: FocusPresetHighCognitiveLoad, want: "Low"},
		{name: "medium wraps", start: FocusPresetMedium, next: FocusPresetHigh, previous: FocusPresetLow, want: "Medium"},
		{name: "high wraps", start: FocusPresetHigh, next: FocusPresetHighCognitiveLoad, previous: FocusPresetMedium, want: "High"},
		{name: "cognitive wraps", start: FocusPresetHighCognitiveLoad, next: FocusPresetLow, previous: FocusPresetHigh, want: "High cognitive load"},
		{name: "unknown falls back", start: FocusPreset(99), next: FocusPresetMedium, previous: FocusPresetMedium, want: "Unknown"},
	}

	for _, test := range tests {
		if got := test.start.Next(); got != test.next {
			t.Fatalf("%s next = %v, want %v", test.name, got, test.next)
		}
		if got := test.start.Previous(); got != test.previous {
			t.Fatalf("%s previous = %v, want %v", test.name, got, test.previous)
		}
		if got := test.start.String(); got != test.want {
			t.Fatalf("%s string = %q, want %q", test.name, got, test.want)
		}
	}
}

func TestADHDPresetCycle(t *testing.T) {
	tests := []struct {
		name     string
		start    ADHDPreset
		next     ADHDPreset
		previous ADHDPreset
		want     string
	}{
		{name: "white wraps", start: ADHDPresetWhite, next: ADHDPresetPink, previous: ADHDPresetPink, want: "White"},
		{name: "pink wraps", start: ADHDPresetPink, next: ADHDPresetWhite, previous: ADHDPresetWhite, want: "Pink"},
		{name: "unknown falls back", start: ADHDPreset(99), next: ADHDPresetWhite, previous: ADHDPresetWhite, want: "Unknown"},
	}

	for _, test := range tests {
		if got := test.start.Next(); got != test.next {
			t.Fatalf("%s next = %v, want %v", test.name, got, test.next)
		}
		if got := test.start.Previous(); got != test.previous {
			t.Fatalf("%s previous = %v, want %v", test.name, got, test.previous)
		}
		if got := test.start.String(); got != test.want {
			t.Fatalf("%s string = %q, want %q", test.name, got, test.want)
		}
	}
}

func TestPausedState(t *testing.T) {
	generator := NewGenerator()

	if generator.Paused() {
		t.Fatal("generator starts paused, want playing")
	}

	generator.SetPaused(true)
	if !generator.Paused() {
		t.Fatal("generator paused state = false, want true")
	}

	generator.TogglePaused()
	if generator.Paused() {
		t.Fatal("generator paused state = true after toggle, want false")
	}
}

func TestModeGainCompensationOrdering(t *testing.T) {
	brown := modeGain(ModeBrown)
	focus := modeGain(ModeFocus)
	adhdWhite := adhdPresetGain(ADHDPresetWhite)
	adhdPink := adhdPresetGain(ADHDPresetPink)
	pink := modeGain(ModePink)
	speech := modeGain(ModeSpeech)

	if !(brown > focus) {
		t.Fatalf("brown gain = %.2f, want > focus gain %.2f", brown, focus)
	}
	if !(focus > adhdPink) {
		t.Fatalf("focus gain = %.2f, want > adhd pink gain %.2f", focus, adhdPink)
	}
	if !(focus > pink) {
		t.Fatalf("focus gain = %.2f, want > pink gain %.2f", focus, pink)
	}
	if !(brown > pink) {
		t.Fatalf("brown gain = %.2f, want > pink gain %.2f", brown, pink)
	}
	if !(adhdPink > adhdWhite) {
		t.Fatalf("adhd pink gain = %.2f, want > adhd white gain %.2f", adhdPink, adhdWhite)
	}
	if !(adhdWhite > speech) {
		t.Fatalf("adhd white gain = %.2f, want > speech gain %.2f", adhdWhite, speech)
	}
	if !(pink > speech) {
		t.Fatalf("pink gain = %.2f, want > speech gain %.2f", pink, speech)
	}
}

func TestModeGainCompensationValues(t *testing.T) {
	if got := modeGain(ModeFocus); got != 1.80 {
		t.Fatalf("focus gain = %.2f, want 1.80", got)
	}
	if got := modeGain(ModeBrown); got != 4.10 {
		t.Fatalf("brown gain = %.2f, want 4.10", got)
	}
	if got := modeGain(ModePink); got != 1.00 {
		t.Fatalf("pink gain = %.2f, want 1.00", got)
	}
	if got := modeGain(ModeSpeech); got != 0.24 {
		t.Fatalf("speech gain = %.2f, want 0.24", got)
	}
	if got := adhdPresetGain(ADHDPresetWhite); got != 0.42 {
		t.Fatalf("adhd white gain = %.2f, want 0.42", got)
	}
	if got := adhdPresetGain(ADHDPresetPink); got != 1.00 {
		t.Fatalf("adhd pink gain = %.2f, want 1.00", got)
	}
}

func TestADHDPinkGainMatchesStandardPink(t *testing.T) {
	if got, want := adhdPresetGain(ADHDPresetPink), modeGain(ModePink); got != want {
		t.Fatalf("adhd pink gain = %.2f, want standard pink gain %.2f", got, want)
	}
}

func TestFocusDensityAddsLayers(t *testing.T) {
	low := focusRMS(FocusPresetLow)
	medium := focusRMS(FocusPresetMedium)
	high := focusRMS(FocusPresetHigh)
	cognitive := focusRMS(FocusPresetHighCognitiveLoad)
	mediumMotion := focusMotionEnergy(FocusPresetMedium)
	highMotion := focusMotionEnergy(FocusPresetHigh)

	if !(medium > low) {
		t.Fatalf("medium rms = %.6f, want > low rms %.6f", medium, low)
	}
	if !(highMotion > mediumMotion) {
		t.Fatalf("high motion = %.6f, want > medium motion %.6f", highMotion, mediumMotion)
	}
	if !(cognitive > low) {
		t.Fatalf("cognitive rms = %.6f, want > low rms %.6f", cognitive, low)
	}
	if !(high > low) {
		t.Fatalf("high rms = %.6f, want > low rms %.6f", high, low)
	}
}

func TestADHDWhitePresetKeepsMoreHighBandEnergyThanPink(t *testing.T) {
	whiteRatio := highBandRatio(ADHDPresetWhite)
	pinkRatio := highBandRatio(ADHDPresetPink)

	if !(whiteRatio > pinkRatio) {
		t.Fatalf("white high-band ratio = %.6f, want > pink high-band ratio %.6f", whiteRatio, pinkRatio)
	}
}

func TestStructuredPulseContourStaysSubtle(t *testing.T) {
	const depth = 0.12

	start := structuredPulseContour(0.0, depth)
	peak := structuredPulseContour(0.18, depth)
	tail := structuredPulseContour(0.95, depth)
	shapeStart := structuredPulseShape(0.0)
	shapePeak := structuredPulseShape(0.18)
	shapeLate := structuredPulseShape(0.95)

	if start != 1.0 {
		t.Fatalf("start contour = %.6f, want 1.0", start)
	}
	if peak <= 1.05 || peak >= 1.15 {
		t.Fatalf("peak contour = %.6f, want subtle 5-15%% lift", peak)
	}
	if tail <= 1.0 || tail >= peak {
		t.Fatalf("tail contour = %.6f, want a long release between 1.0 and peak %.6f", tail, peak)
	}
	if shapeStart != 0.0 {
		t.Fatalf("shape start = %.6f, want 0", shapeStart)
	}
	if shapePeak != 1.0 {
		t.Fatalf("shape peak = %.6f, want 1", shapePeak)
	}
	if shapeLate <= 0.0 || shapeLate >= 1.0 {
		t.Fatalf("shape late = %.6f, want within (0, 1)", shapeLate)
	}
}

func TestHighCognitiveLoadPresetReducesMotionProfile(t *testing.T) {
	medium := focusPresetProfile(FocusPresetMedium)
	high := focusPresetProfile(FocusPresetHigh)
	cognitive := focusPresetProfile(FocusPresetHighCognitiveLoad)

	if cognitive.textureMix != 0 {
		t.Fatalf("cognitive texture mix = %.6f, want 0", cognitive.textureMix)
	}
	if cognitive.bedMix >= medium.bedMix {
		t.Fatalf("cognitive bed mix = %.6f, want < medium bed mix %.6f", cognitive.bedMix, medium.bedMix)
	}
	if cognitive.pulseDepth < 0.05 || cognitive.pulseDepth > 0.15 {
		t.Fatalf("cognitive pulse depth = %.6f, want within 5-15%%", cognitive.pulseDepth)
	}
	if high.pulseDepth < medium.pulseDepth {
		t.Fatalf("high pulse depth = %.6f, want >= medium pulse depth %.6f", high.pulseDepth, medium.pulseDepth)
	}
	if cognitive.bedPulseDepth >= high.bedPulseDepth {
		t.Fatalf("cognitive bed pulse depth = %.6f, want < high bed pulse depth %.6f", cognitive.bedPulseDepth, high.bedPulseDepth)
	}
	for _, test := range []struct {
		name    string
		profile focusProfile
	}{
		{name: "medium", profile: medium},
		{name: "high", profile: high},
		{name: "cognitive", profile: cognitive},
	} {
		ratio := test.profile.bedPulseDepth / test.profile.pulseDepth
		if ratio < 0.399 || ratio > 0.601 {
			t.Fatalf("%s bed-to-pulse ratio = %.6f, want within [0.40, 0.60]", test.name, ratio)
		}
	}
}

func TestFocusStructuredPulseOverlayValidation(t *testing.T) {
	beatSamples := int(math.Round(config.SampleRate * 60.0 / focusTempoBPM))
	derivedTempo := 60.0 * config.SampleRate / float64(beatSamples)
	if math.Abs(derivedTempo-focusTempoBPM) > 0.05 {
		t.Fatalf("derived tempo = %.6f BPM, want %.6f BPM", derivedTempo, focusTempoBPM)
	}

	medium := analyzeFocusPreset(FocusPresetMedium)
	high := analyzeFocusPreset(FocusPresetHigh)
	cognitive := analyzeFocusPreset(FocusPresetHighCognitiveLoad)

	if medium.peakWindowIndex != high.peakWindowIndex {
		t.Fatalf("peak window mismatch: medium=%d high=%d, want one contour cycle aligned per beat", medium.peakWindowIndex, high.peakWindowIndex)
	}
	if medium.peakWindowIndex > 2 {
		t.Fatalf("peak window index = %d, want an early-beat contour peak", medium.peakWindowIndex)
	}
	if medium.meanSwing <= 0 || medium.meanSwing >= 0.20 {
		t.Fatalf("medium mean swing = %.6f, want within (0, 0.20)", medium.meanSwing)
	}
	if cognitive.meanSwing >= high.meanSwing {
		t.Fatalf("cognitive mean swing = %.6f, want < high mean swing %.6f", cognitive.meanSwing, high.meanSwing)
	}
	if high.windowPeakConsistency < 0.99 {
		t.Fatalf("high peak consistency = %.6f, want a single dominant peak window per beat", high.windowPeakConsistency)
	}
	if high.beatRMSCV >= 0.08 {
		t.Fatalf("high beat RMS CV = %.6f, want narrow beat-to-beat loudness swing", high.beatRMSCV)
	}
	if cognitive.beatRMSCV >= 0.08 {
		t.Fatalf("cognitive beat RMS CV = %.6f, want narrow beat-to-beat loudness swing", cognitive.beatRMSCV)
	}
}

func TestTextureDepthInterpolatesAcrossBars(t *testing.T) {
	start := textureDepth(7, 0.0)
	end := textureDepth(7, 1.0)
	mid := textureDepth(7, 0.5)

	wantStart := 0.16 + 0.14*unitHash(7)
	wantEnd := 0.16 + 0.14*unitHash(8)

	if start != wantStart {
		t.Fatalf("start depth = %.6f, want %.6f", start, wantStart)
	}
	if end != wantEnd {
		t.Fatalf("end depth = %.6f, want %.6f", end, wantEnd)
	}

	lower := math.Min(wantStart, wantEnd)
	upper := math.Max(wantStart, wantEnd)
	if mid < lower || mid > upper {
		t.Fatalf("mid depth = %.6f, want within [%.6f, %.6f]", mid, lower, upper)
	}
}

func TestFillWritesStereoSamples(t *testing.T) {
	generator := NewGenerator()
	samples := make([]float32, 16)

	generator.Fill(samples)

	for i := 0; i < len(samples); i += 2 {
		if samples[i] < -config.MaxVolume || samples[i] > config.MaxVolume {
			t.Fatalf("left sample %d out of range: %f", i, samples[i])
		}
		if samples[i+1] < -config.MaxVolume || samples[i+1] > config.MaxVolume {
			t.Fatalf("right sample %d out of range: %f", i+1, samples[i+1])
		}
	}
}

func TestFillWritesSilenceWhenPaused(t *testing.T) {
	generator := NewGenerator()
	samples := make([]float32, 16)
	for i := range samples {
		samples[i] = 1
	}

	generator.SetPaused(true)
	generator.Fill(samples)

	for i, sample := range samples {
		if sample != 0 {
			t.Fatalf("sample %d = %f, want 0", i, sample)
		}
	}
}

func TestSpeechShaperEmphasizesSpeechBand(t *testing.T) {
	low := steadyStateGain(NewSpeechShaper(), 160)
	mid := steadyStateGain(NewSpeechShaper(), 1000)
	high := steadyStateGain(NewSpeechShaper(), 6000)

	if mid <= low*1.30 {
		t.Fatalf("mid-band gain = %.6f, want > %.6f", mid, low*1.30)
	}
	if mid <= high*1.50 {
		t.Fatalf("mid-band gain = %.6f, want > %.6f", mid, high*1.50)
	}
}

func steadyStateGain(shaper SpeechShaper, frequencyHz float64) float64 {
	const amplitude = 0.25
	totalSamples := config.SampleRate * 2
	settleSamples := config.SampleRate / 2

	var inputEnergy float64
	var outputEnergy float64

	for i := 0; i < totalSamples; i++ {
		phase := 2 * math.Pi * frequencyHz * float64(i) / config.SampleRate
		input := float32(amplitude * math.Sin(phase))
		output := shaper.Process(input)
		if i < settleSamples {
			continue
		}

		inputEnergy += float64(input * input)
		outputEnergy += float64(output * output)
	}

	return math.Sqrt(outputEnergy / inputEnergy)
}

func focusRMS(preset FocusPreset) float64 {
	rng := xorShift32{x: 0x12345678}
	state := NewFocusState()
	sampleCount := config.SampleRate * 3
	settleSamples := config.SampleRate / 2

	var energy float64
	var counted int

	for i := 0; i < sampleCount; i++ {
		left, right := state.NextPair(&rng, preset)
		if i < settleSamples {
			continue
		}

		energy += float64(left*left + right*right)
		counted += 2
	}

	return math.Sqrt(energy / float64(counted))
}

func highBandRatio(preset ADHDPreset) float64 {
	generator := NewGenerator()
	generator.SetMode(ModeADHD)
	generator.SetADHDPreset(preset)

	const frameCount = config.SampleRate * 2
	samples := make([]float32, frameCount*2)
	generator.Fill(samples)

	highpass := NewOnePoleHP(2000)
	lowpass := NewOnePoleLP(300)

	var highEnergy float64
	var lowEnergy float64
	for i := 0; i < len(samples); i += 2 {
		mid := 0.5 * (samples[i] + samples[i+1])
		high := highpass.Process(mid)
		low := lowpass.Process(mid)
		highEnergy += float64(high * high)
		lowEnergy += float64(low * low)
	}

	if lowEnergy == 0 {
		return math.Inf(1)
	}

	return highEnergy / lowEnergy
}

func focusMotionEnergy(preset FocusPreset) float64 {
	rng := xorShift32{x: 0x12345678}
	state := NewFocusState()
	sampleCount := config.SampleRate * 3
	settleSamples := config.SampleRate / 2

	var energy float64
	var counted int
	var prevLeft float32
	var prevRight float32
	var havePrevious bool

	for i := 0; i < sampleCount; i++ {
		left, right := state.NextPair(&rng, preset)
		if i < settleSamples {
			prevLeft = left
			prevRight = right
			havePrevious = true
			continue
		}
		if !havePrevious {
			prevLeft = left
			prevRight = right
			havePrevious = true
			continue
		}

		deltaLeft := left - prevLeft
		deltaRight := right - prevRight
		energy += float64(deltaLeft*deltaLeft + deltaRight*deltaRight)
		counted += 2

		prevLeft = left
		prevRight = right
	}

	return math.Sqrt(energy / float64(counted))
}

type focusBeatAnalysis struct {
	meanSwing             float64
	beatRMSCV             float64
	peakWindowIndex       int
	windowPeakConsistency float64
}

func analyzeFocusPreset(preset FocusPreset) focusBeatAnalysis {
	rng := xorShift32{x: 0x12345678}
	state := NewFocusState()
	beatSamples := int(math.Round(config.SampleRate * 60.0 / focusTempoBPM))
	windowCount := 8
	windowSamples := beatSamples / windowCount
	settleBeats := 2
	measuredBeats := 8
	totalSamples := (settleBeats + measuredBeats) * beatSamples

	swings := make([]float64, 0, measuredBeats)
	beatRMS := make([]float64, 0, measuredBeats)
	peakCounts := make([]int, windowCount)

	for beat := 0; beat < settleBeats+measuredBeats; beat++ {
		windowEnergy := make([]float64, windowCount)
		windowSamplesCount := make([]int, windowCount)
		var beatEnergy float64
		var beatCount int

		for i := 0; i < beatSamples && beat*beatSamples+i < totalSamples; i++ {
			left, right := state.NextPair(&rng, preset)
			energy := float64(left*left+right*right) / 2.0
			windowIndex := i / windowSamples
			if windowIndex >= windowCount {
				windowIndex = windowCount - 1
			}

			windowEnergy[windowIndex] += energy
			windowSamplesCount[windowIndex]++
			beatEnergy += energy
			beatCount++
		}

		if beat < settleBeats {
			continue
		}

		windowRMS := make([]float64, windowCount)
		minRMS := math.MaxFloat64
		maxRMS := 0.0
		peakIndex := 0

		for i := range windowEnergy {
			windowRMS[i] = math.Sqrt(windowEnergy[i] / float64(windowSamplesCount[i]))
			if windowRMS[i] < minRMS {
				minRMS = windowRMS[i]
			}
			if windowRMS[i] > maxRMS {
				maxRMS = windowRMS[i]
				peakIndex = i
			}
		}

		peakCounts[peakIndex]++
		swings = append(swings, (maxRMS-minRMS)/maxRMS)
		beatRMS = append(beatRMS, math.Sqrt(beatEnergy/float64(beatCount)))
	}

	peakIndex := 0
	peakCount := 0
	for i, count := range peakCounts {
		if count > peakCount {
			peakCount = count
			peakIndex = i
		}
	}

	return focusBeatAnalysis{
		meanSwing:             mean(swings),
		beatRMSCV:             coefficientOfVariation(beatRMS),
		peakWindowIndex:       peakIndex,
		windowPeakConsistency: float64(peakCount) / float64(measuredBeats),
	}
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var total float64
	for _, value := range values {
		total += value
	}

	return total / float64(len(values))
}

func coefficientOfVariation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	m := mean(values)
	if m == 0 {
		return 0
	}

	var variance float64
	for _, value := range values {
		delta := value - m
		variance += delta * delta
	}

	return math.Sqrt(variance/float64(len(values))) / m
}
