package audio

type SampleFiller interface {
	Fill(samples []float32)
}
