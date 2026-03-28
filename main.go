package main

/*
#cgo CFLAGS: -Wno-deprecated-declarations
#cgo LDFLAGS: -framework AudioToolbox -framework CoreAudio -framework Foundation
#include <AudioToolbox/AudioToolbox.h>

extern void AQOutputCallback(void *inUserData, AudioQueueRef inAQ, AudioQueueBufferRef inBuffer);
*/
import "C"

import (
	"fmt"
	"log"
	"math"
	"sync/atomic"
	"unsafe"

	"github.com/getlantern/systray"
)

const (
	sampleRate = 48000
	channels   = 2
	frames     = 512
	numBuffers = 3

	defaultVolume = 0.025
	minVolume     = 0.0
	maxVolume     = 0.08
	volumeStep    = 0.005
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

type xorShift32 struct {
	x uint32
}

func (r *xorShift32) nextFloat32() float32 {
	x := r.x
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	r.x = x
	u := float32(x) / float32(math.MaxUint32)
	return 2*u - 1
}

type PinkState struct {
	b0, b1, b2, b3, b4, b5, b6 float32
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
	dt := 1.0 / sampleRate
	alpha := float32(rc / (rc + dt))
	return OnePoleHP{alpha: alpha}
}

func (f *OnePoleHP) Process(x float32) float32 {
	y := f.alpha * (f.yPrev + x - f.xPrev)
	f.xPrev = x
	f.yPrev = y
	return y
}

type OnePoleLP struct {
	alpha float32
	y     float32
}

func NewOnePoleLP(cutoffHz float32) OnePoleLP {
	dt := 1.0 / sampleRate
	rc := 1.0 / (2.0 * math.Pi * float64(cutoffHz))
	alpha := float32(dt / (rc + dt))
	return OnePoleLP{alpha: alpha}
}

func (f *OnePoleLP) Process(x float32) float32 {
	f.y += f.alpha * (x - f.y)
	return f.y
}

type Generator struct {
	rng xorShift32

	modeBits   atomic.Int32
	volumeBits atomic.Uint32

	brownL float32
	brownR float32

	pinkL PinkState
	pinkR PinkState

	voiceHP_L OnePoleHP
	voiceHP_R OnePoleHP
	voiceLP_L OnePoleLP
	voiceLP_R OnePoleLP
}

func NewGenerator() *Generator {
	g := &Generator{
		rng: xorShift32{x: 0x12345678},

		voiceHP_L: NewOnePoleHP(700),
		voiceHP_R: NewOnePoleHP(700),
		voiceLP_L: NewOnePoleLP(3500),
		voiceLP_R: NewOnePoleLP(3500),
	}
	g.SetMode(ModeBrown)
	g.SetVolume(defaultVolume)
	return g
}

func (g *Generator) Mode() Mode {
	return Mode(g.modeBits.Load())
}

func (g *Generator) SetMode(m Mode) {
	g.modeBits.Store(int32(m))
}

func (g *Generator) Volume() float32 {
	return math.Float32frombits(g.volumeBits.Load())
}

func (g *Generator) SetVolume(v float32) {
	if v < minVolume {
		v = minVolume
	}
	if v > maxVolume {
		v = maxVolume
	}
	g.volumeBits.Store(math.Float32bits(v))
}

func clamp(x float32) float32 {
	if x > 1 {
		return 1
	}
	if x < -1 {
		return -1
	}
	return x
}

func (g *Generator) nextBrownPair() (float32, float32) {
	wL := g.rng.nextFloat32()
	wR := g.rng.nextFloat32()

	g.brownL = 0.995*g.brownL + 0.02*wL
	g.brownR = 0.995*g.brownR + 0.02*wR

	return clamp(g.brownL), clamp(g.brownR)
}

func (g *Generator) nextPinkPair() (float32, float32) {
	wL := g.rng.nextFloat32()
	wR := g.rng.nextFloat32()
	return clamp(g.pinkL.Next(wL)), clamp(g.pinkR.Next(wR))
}

func (g *Generator) nextVoicePair() (float32, float32) {
	wL := g.rng.nextFloat32()
	wR := g.rng.nextFloat32()

	xL := g.pinkL.Next(wL)
	xR := g.pinkR.Next(wR)

	xL = g.voiceHP_L.Process(xL)
	xR = g.voiceHP_R.Process(xR)

	xL = g.voiceLP_L.Process(xL)
	xR = g.voiceLP_R.Process(xR)

	return clamp(1.8 * xL), clamp(1.8 * xR)
}

func (g *Generator) Fill(samples []float32) {
	vol := g.Volume()
	mode := g.Mode()

	for i := 0; i < len(samples); i += 2 {
		var l, r float32
		switch mode {
		case ModeBrown:
			l, r = g.nextBrownPair()
		case ModePink:
			l, r = g.nextPinkPair()
		case ModeVoice:
			l, r = g.nextVoicePair()
		}
		samples[i] = vol * l
		samples[i+1] = vol * r
	}
}

var currentGen *Generator
var audioQueue C.AudioQueueRef

func fillAudioBuffer(buf *C.AudioQueueBuffer) {
	if currentGen == nil {
		return
	}
	sampleCount := int(buf.mAudioDataBytesCapacity) / 4
	samples := unsafe.Slice((*float32)(unsafe.Pointer(buf.mAudioData)), sampleCount)
	currentGen.Fill(samples)
	buf.mAudioDataByteSize = C.UInt32(sampleCount * 4)
}

//export AQOutputCallback
func AQOutputCallback(inUserData unsafe.Pointer, inAQ C.AudioQueueRef, inBuffer C.AudioQueueBufferRef) {
	fillAudioBuffer(inBuffer)
	C.AudioQueueEnqueueBuffer(inAQ, inBuffer, 0, nil)
}

func checkStatus(name string, status C.OSStatus) {
	if status != 0 {
		log.Fatalf("%s failed: OSStatus=%d", name, int32(status))
	}
}

func startAudio() {
	currentGen = NewGenerator()

	var format C.AudioStreamBasicDescription
	var runLoop C.CFRunLoopRef
	var runMode C.CFStringRef
	format.mSampleRate = C.Float64(sampleRate)
	format.mFormatID = C.kAudioFormatLinearPCM
	format.mFormatFlags = C.kLinearPCMFormatFlagIsFloat | C.kAudioFormatFlagIsPacked
	format.mBytesPerPacket = C.UInt32(channels * 4)
	format.mFramesPerPacket = 1
	format.mBytesPerFrame = C.UInt32(channels * 4)
	format.mChannelsPerFrame = channels
	format.mBitsPerChannel = 32

	checkStatus("AudioQueueNewOutput", C.AudioQueueNewOutput(
		&format,
		(C.AudioQueueOutputCallback)(unsafe.Pointer(C.AQOutputCallback)),
		nil,
		runLoop,
		runMode,
		0,
		&audioQueue,
	))

	bufferBytes := C.UInt32(frames * channels * 4)
	for i := 0; i < numBuffers; i++ {
		var buf C.AudioQueueBufferRef
		checkStatus("AudioQueueAllocateBuffer", C.AudioQueueAllocateBuffer(audioQueue, bufferBytes, &buf))
		fillAudioBuffer(buf)
		checkStatus("AudioQueueEnqueueBuffer", C.AudioQueueEnqueueBuffer(audioQueue, buf, 0, nil))
	}

	checkStatus("AudioQueueStart", C.AudioQueueStart(audioQueue, nil))
}

func stopAudio() {
	if audioQueue != nil {
		C.AudioQueueStop(audioQueue, 1)
		C.AudioQueueDispose(audioQueue, 1)
		audioQueue = nil
	}
}

func updateChecks(mBrown, mPink, mVoice *systray.MenuItem) {
	mode := currentGen.Mode()
	mBrown.Uncheck()
	mPink.Uncheck()
	mVoice.Uncheck()

	switch mode {
	case ModeBrown:
		mBrown.Check()
	case ModePink:
		mPink.Check()
	case ModeVoice:
		mVoice.Check()
	}
}

func statusText() string {
	return fmt.Sprintf("Mode: %s | Vol: %.3f", currentGen.Mode(), currentGen.Volume())
}

func onReady() {
	// You can replace these with your own icon bytes if you want.
	systray.SetTitle("Mask")
	systray.SetTooltip("Noise masking")

	startAudio()

	mStatus := systray.AddMenuItem(statusText(), "Current masking status")
	mStatus.Disable()

	systray.AddSeparator()

	mBrown := systray.AddMenuItemCheckbox("Brown", "Low rumble / HVAC / travel", true)
	mPink := systray.AddMenuItemCheckbox("Pink", "General ambient masking", false)
	mVoice := systray.AddMenuItemCheckbox("Voice-focused", "Reduce speech intelligibility", false)

	systray.AddSeparator()

	mVolUp := systray.AddMenuItem("Volume +", "Increase masker volume")
	mVolDown := systray.AddMenuItem("Volume -", "Decrease masker volume")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit the app")

	updateChecks(mBrown, mPink, mVoice)

	go func() {
		for {
			select {
			case <-mBrown.ClickedCh:
				currentGen.SetMode(ModeBrown)
				updateChecks(mBrown, mPink, mVoice)
				mStatus.SetTitle(statusText())

			case <-mPink.ClickedCh:
				currentGen.SetMode(ModePink)
				updateChecks(mBrown, mPink, mVoice)
				mStatus.SetTitle(statusText())

			case <-mVoice.ClickedCh:
				currentGen.SetMode(ModeVoice)
				updateChecks(mBrown, mPink, mVoice)
				mStatus.SetTitle(statusText())

			case <-mVolUp.ClickedCh:
				currentGen.SetVolume(currentGen.Volume() + volumeStep)
				mStatus.SetTitle(statusText())

			case <-mVolDown.ClickedCh:
				currentGen.SetVolume(currentGen.Volume() - volumeStep)
				mStatus.SetTitle(statusText())

			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	stopAudio()
}

func main() {
	systray.Run(onReady, onExit)
}
