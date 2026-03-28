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
		{name: "focus wraps", start: ModeFocus, next: ModeBrown, previous: ModeSpeech},
		{name: "brown wraps", start: ModeBrown, next: ModePink, previous: ModeFocus},
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
	pink := modeGain(ModePink)
	speech := modeGain(ModeSpeech)

	if !(brown > focus) {
		t.Fatalf("brown gain = %.2f, want > focus gain %.2f", brown, focus)
	}
	if !(focus > pink) {
		t.Fatalf("focus gain = %.2f, want > pink gain %.2f", focus, pink)
	}
	if !(brown > pink) {
		t.Fatalf("brown gain = %.2f, want > pink gain %.2f", brown, pink)
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

func TestStructuredPulseContourStaysSubtle(t *testing.T) {
	const depth = 0.12

	start := structuredPulseContour(0.0, depth)
	peak := structuredPulseContour(0.18, depth)
	tail := structuredPulseContour(0.95, depth)

	if start != 1.0 {
		t.Fatalf("start contour = %.6f, want 1.0", start)
	}
	if peak <= 1.05 || peak >= 1.15 {
		t.Fatalf("peak contour = %.6f, want subtle 5-15%% lift", peak)
	}
	if tail <= 1.0 || tail >= peak {
		t.Fatalf("tail contour = %.6f, want a long release between 1.0 and peak %.6f", tail, peak)
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
