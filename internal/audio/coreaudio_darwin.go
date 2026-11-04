//go:build darwin

package audio

/*
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/gen2brain/malgo"
)

// systemSources lists Core Audio capture devices. Missing system-output
// loopback is not an error and is not reported as a source. Enumeration
// does not request microphone permission; that happens when capture starts.
func systemSources() ([]Source, error) {
	ctx, err := newCoreContext()
	if err != nil {
		return nil, err
	}
	defer releaseContext(ctx)
	infos, err := ctx.Devices(malgo.Capture)
	if err != nil {
		return nil, wrapCapture(err)
	}
	devs := make([]inputDevice, 0, len(infos))
	for _, info := range infos {
		uid := deviceUID(info.ID[:])
		if uid == "" {
			continue
		}
		devs = append(devs, inputDevice{
			ID:      uid,
			Name:    info.Name(),
			Default: info.IsDefault != 0,
		})
	}
	return micSources(devs), nil
}

// captureSystem records one Core Audio input as 16 kHz mono float32.
// sourceName is the device UID from systemSources, without the "source:" prefix.
func captureSystem(ctx context.Context, sourceName string) (<-chan []float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := ensureMicPermission(ctx); err != nil {
		return nil, err
	}
	id, err := deviceIDFromUID(sourceName)
	if err != nil {
		return nil, err
	}
	actx, err := newCoreContext()
	if err != nil {
		return nil, err
	}
	cfg := malgo.DefaultDeviceConfig(malgo.Capture)
	cfg.SampleRate = uint32(SampleRate)
	cfg.PeriodSizeInFrames = uint32(SampleRate / 10)
	cfg.PeriodSizeInMilliseconds = 0
	cfg.Capture.Format = malgo.FormatF32
	cfg.Capture.Channels = 1
	// miniaudio copies this id inside ma_device_init, so the allocation
	// only has to live until InitDevice returns.
	ptr := id.Pointer()
	cfg.Capture.DeviceID = ptr

	stream := newMicStream()
	device, err := malgo.InitDevice(actx.Context, cfg, malgo.DeviceCallbacks{
		Data: func(_, input []byte, _ uint32) {
			stream.push(input)
		},
	})
	C.free(ptr)
	if err != nil {
		releaseContext(actx)
		if micPermissionBlocked() {
			return nil, microphoneDenied()
		}
		return nil, wrapCapture(err)
	}
	sampleBytes, err := captureSampleBytes(device.CaptureFormat())
	if err != nil {
		device.Uninit()
		releaseContext(actx)
		return nil, err
	}
	channels := int(device.CaptureChannels())
	if channels < 1 {
		channels = 1
	}
	rate := int(device.SampleRate())
	if rate < 1 {
		rate = SampleRate
	}
	stream.configure(sampleBytes, channels, rate)
	if err := device.Start(); err != nil {
		device.Uninit()
		releaseContext(actx)
		if micPermissionBlocked() {
			return nil, microphoneDenied()
		}
		return nil, wrapCapture(err)
	}
	go func() {
		<-ctx.Done()
		_ = device.Stop()
		device.Uninit()
		stream.close()
		releaseContext(actx)
	}()
	return stream.out, nil
}

func newCoreContext() (*malgo.AllocatedContext, error) {
	// Only Core Audio. The default backend list can fall through to the
	// null device, which would look like a live microphone.
	ctx, err := malgo.InitContext([]malgo.Backend{malgo.BackendCoreaudio}, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, wrapCapture(err)
	}
	return ctx, nil
}

func releaseContext(ctx *malgo.AllocatedContext) {
	if ctx == nil {
		return
	}
	_ = ctx.Uninit()
	ctx.Free()
}

func deviceIDFromUID(uid string) (malgo.DeviceID, error) {
	var id malgo.DeviceID
	if err := putDeviceUID(id[:], uid); err != nil {
		return malgo.DeviceID{}, wrapCapture(err)
	}
	return id, nil
}

func captureSampleBytes(format malgo.FormatType) (int, error) {
	switch format {
	case malgo.FormatF32:
		return 4, nil
	case malgo.FormatS16:
		return 2, nil
	default:
		return 0, wrapCapture(errUnsupportedFormat)
	}
}

var errUnsupportedFormat = errString("microphone format is not f32 or s16")

type errString string

func (e errString) Error() string { return string(e) }

// micStream copies callback PCM into a buffered channel. The callback
// runs on Core Audio's realtime thread, so a full queue drops the oldest
// chunk instead of blocking.
type micStream struct {
	mu          sync.Mutex
	closed      bool
	out         chan []float32
	sampleBytes atomic.Int32
	channels    atomic.Int32
	rate        atomic.Int32
}

func newMicStream() *micStream {
	s := &micStream{out: make(chan []float32, 64)}
	s.sampleBytes.Store(4)
	s.channels.Store(1)
	s.rate.Store(SampleRate)
	return s
}

func (s *micStream) configure(sampleBytes, channels, rate int) {
	s.sampleBytes.Store(int32(sampleBytes))
	s.channels.Store(int32(channels))
	s.rate.Store(int32(rate))
}

func (s *micStream) push(input []byte) {
	if len(input) == 0 {
		return
	}
	samples := float32FromPCM(input, int(s.sampleBytes.Load()), int(s.channels.Load()))
	rate := int(s.rate.Load())
	if rate > 0 && rate != SampleRate {
		samples = Resample(samples, rate, SampleRate)
	}
	if len(samples) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	select {
	case s.out <- samples:
	default:
		select {
		case <-s.out:
		default:
		}
		select {
		case s.out <- samples:
		default:
		}
	}
}

func (s *micStream) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.out)
}
