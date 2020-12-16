package publish

import (
	"errors"
	"fmt"
	"github.com/ebml-go/webm"
	"log"
	"os"
	"time"
)

type WebmAudio struct {
	file        *os.File
	reader      *webm.Reader
	lastGranule uint64

	channels   uint
	sampleRate float64
}

func NewWebmAudio(audioFileName string) (*WebmAudio, error) {
	file, err := os.Open(audioFileName)
	if err != nil {
		return nil, err
	}

	var w webm.WebM
	reader, err := webm.Parse(file, &w)
	if err != nil {
		return nil, err
	}
	track := w.FindFirstAudioTrack()
	log.Println("Duration:", w.Segment.GetDuration())

	if track.IsAudio() == false {
		return nil, errors.New("this file doesn't have audio track")
	}

	return &WebmAudio{file: file, reader: reader, lastGranule: 0, channels: track.Channels, sampleRate: track.SamplingFrequency}, nil
}

func (o *WebmAudio) Next() ([]byte, float64, error) {
	o.reader.Seek(time.Millisecond * 100)
	packet := <-o.reader.Chan
	fmt.Println(packet)

	return packet.Data, packet.Timecode.Seconds() * o.sampleRate, nil
}

func (o *WebmAudio) Reset() error {
	return nil
}

func (o WebmAudio) Close() error {
	return o.file.Close()
}

func (o WebmAudio) GetSampleRate() float64 {
	return o.sampleRate
}

func (o WebmAudio) GetLastGranule() uint64 {
	return o.lastGranule
}

func (o WebmAudio) GetSleepDuration(sampleCount float64) time.Duration {
	return time.Duration((sampleCount/o.GetSampleRate())*1000) * time.Millisecond
}
