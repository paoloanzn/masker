//go:build darwin

package audio

/*
#cgo CFLAGS: -Wno-deprecated-declarations
#cgo LDFLAGS: -framework AudioToolbox -framework CoreAudio -framework Foundation
#include <AudioToolbox/AudioToolbox.h>

extern void AQOutputCallback(void *inUserData, AudioQueueRef inAQ, AudioQueueBufferRef inBuffer);
*/
import "C"

import (
	"fmt"
	"unsafe"

	"masker/internal/config"
)

type Engine struct {
	filler SampleFiller
	queue  C.AudioQueueRef
}

var currentEngine *Engine

func NewEngine(filler SampleFiller) *Engine {
	return &Engine{filler: filler}
}

func (e *Engine) Start() error {
	currentEngine = e

	var format C.AudioStreamBasicDescription
	var runLoop C.CFRunLoopRef
	var runMode C.CFStringRef

	format.mSampleRate = C.Float64(config.SampleRate)
	format.mFormatID = C.kAudioFormatLinearPCM
	format.mFormatFlags = C.kLinearPCMFormatFlagIsFloat | C.kAudioFormatFlagIsPacked
	format.mBytesPerPacket = C.UInt32(config.Channels * 4)
	format.mFramesPerPacket = 1
	format.mBytesPerFrame = C.UInt32(config.Channels * 4)
	format.mChannelsPerFrame = config.Channels
	format.mBitsPerChannel = 32

	if err := checkStatus("AudioQueueNewOutput", C.AudioQueueNewOutput(
		&format,
		(C.AudioQueueOutputCallback)(unsafe.Pointer(C.AQOutputCallback)),
		nil,
		runLoop,
		runMode,
		0,
		&e.queue,
	)); err != nil {
		currentEngine = nil
		return err
	}

	bufferBytes := C.UInt32(config.FramesPerBuffer * config.Channels * 4)
	for i := 0; i < config.NumBuffers; i++ {
		var buffer C.AudioQueueBufferRef
		if err := checkStatus("AudioQueueAllocateBuffer", C.AudioQueueAllocateBuffer(e.queue, bufferBytes, &buffer)); err != nil {
			e.Stop()
			return err
		}
		fillAudioBuffer(buffer)
		if err := checkStatus("AudioQueueEnqueueBuffer", C.AudioQueueEnqueueBuffer(e.queue, buffer, 0, nil)); err != nil {
			e.Stop()
			return err
		}
	}

	if err := checkStatus("AudioQueueStart", C.AudioQueueStart(e.queue, nil)); err != nil {
		e.Stop()
		return err
	}

	return nil
}

func (e *Engine) Stop() {
	if e.queue != nil {
		C.AudioQueueStop(e.queue, 1)
		C.AudioQueueDispose(e.queue, 1)
		e.queue = nil
	}
	if currentEngine == e {
		currentEngine = nil
	}
}

func fillAudioBuffer(buffer *C.AudioQueueBuffer) {
	if currentEngine == nil || currentEngine.filler == nil {
		return
	}

	sampleCount := int(buffer.mAudioDataBytesCapacity) / 4
	samples := unsafe.Slice((*float32)(unsafe.Pointer(buffer.mAudioData)), sampleCount)
	currentEngine.filler.Fill(samples)
	buffer.mAudioDataByteSize = C.UInt32(sampleCount * 4)
}

//export AQOutputCallback
func AQOutputCallback(inUserData unsafe.Pointer, inAQ C.AudioQueueRef, inBuffer C.AudioQueueBufferRef) {
	fillAudioBuffer(inBuffer)
	C.AudioQueueEnqueueBuffer(inAQ, inBuffer, 0, nil)
}

func checkStatus(name string, status C.OSStatus) error {
	if status != 0 {
		return fmt.Errorf("%s failed: OSStatus=%d", name, int32(status))
	}
	return nil
}
