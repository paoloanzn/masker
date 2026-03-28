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

func TestModeGainCompensationOrdering(t *testing.T) {
	brown := modeGain(ModeBrown)
	pink := modeGain(ModePink)
	speech := modeGain(ModeSpeech)

	if !(brown > pink) {
		t.Fatalf("brown gain = %.2f, want > pink gain %.2f", brown, pink)
	}
	if !(pink > speech) {
		t.Fatalf("pink gain = %.2f, want > speech gain %.2f", pink, speech)
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
