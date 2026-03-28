package noise

import (
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
		{mode: ModeVoice, want: "Voice-focused"},
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
