package core

import "fmt"

// Video is a playback surface. The fake host advances a clock.
// It does not open a decoder, a file, or a network stream.
type Video interface {
	Play() error
	Pause() error
	Playing() bool
	Seek(sec float64) error
	Position() float64
	Duration() float64
	SetVolume(v float64) error
	Volume() float64
	Tick(dt float64) error
	Close() error
	OnEnded(func())
}

// VideoHost opens a clip by name and a caller-supplied duration.
type VideoHost interface {
	Open(source string, duration float64) (Video, error)
}

// FakeVideoHost counts opens and does not decode the source.
type FakeVideoHost struct {
	Opened int
}

// Open rejects an empty name or a non-positive duration.
func (h *FakeVideoHost) Open(source string, duration float64) (Video, error) {
	if h == nil {
		return nil, fmt.Errorf("%w: nil host", ErrState)
	}
	if source == "" || duration <= 0 {
		return nil, fmt.Errorf("%w: source", ErrState)
	}
	h.Opened++
	return &fakeVideo{source: source, duration: duration, volume: 1}, nil
}

type fakeVideo struct {
	source   string
	duration float64
	position float64
	volume   float64
	playing  bool
	ended    bool
	dead     bool
	onEnded  func()
}

func (v *fakeVideo) Play() error {
	if err := v.alive(); err != nil {
		return err
	}
	if v.position >= v.duration {
		v.position = v.duration
		return nil
	}
	v.playing = true
	return nil
}

func (v *fakeVideo) Pause() error {
	if err := v.alive(); err != nil {
		return err
	}
	v.playing = false
	return nil
}

func (v *fakeVideo) Playing() bool { return v.playing && !v.dead }

func (v *fakeVideo) Seek(sec float64) error {
	if err := v.alive(); err != nil {
		return err
	}
	if sec < 0 || sec > v.duration {
		return fmt.Errorf("%w: seek", ErrRange)
	}
	v.position = sec
	if sec < v.duration {
		v.ended = false
	}
	return nil
}

func (v *fakeVideo) Position() float64 { return v.position }
func (v *fakeVideo) Duration() float64 { return v.duration }
func (v *fakeVideo) Volume() float64   { return v.volume }

func (v *fakeVideo) SetVolume(vol float64) error {
	if err := v.alive(); err != nil {
		return err
	}
	if vol < 0 || vol > 1 {
		return fmt.Errorf("%w: volume", ErrRange)
	}
	v.volume = vol
	return nil
}

func (v *fakeVideo) Tick(dt float64) error {
	if err := v.alive(); err != nil {
		return err
	}
	if dt < 0 {
		return fmt.Errorf("%w: tick", ErrRange)
	}
	if !v.playing {
		return nil
	}
	v.position += dt
	if v.position >= v.duration {
		v.position = v.duration
		v.playing = false
		if !v.ended {
			v.ended = true
			if v.onEnded != nil {
				v.onEnded()
			}
		}
	}
	return nil
}

func (v *fakeVideo) Close() error {
	if v.dead {
		return ErrClosed
	}
	v.dead = true
	v.playing = false
	return nil
}

func (v *fakeVideo) OnEnded(fn func()) { v.onEnded = fn }

func (v *fakeVideo) alive() error {
	if v.dead {
		return ErrClosed
	}
	return nil
}
