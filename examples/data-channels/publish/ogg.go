package publish

import (
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
	"io"
	"os"
	"time"
)

type OggAudio struct {
	file        *os.File
	reader      *oggreader.OggReader
	lastGranule uint64

	channels   uint8
	sampleRate uint32
}

func NewOggAudio(audioFileName string) (*OggAudio, error) {
	file, err := os.Open(audioFileName)
	if err != nil {
		return nil, err
	}

	ogg, h, err := oggreader.NewWith(file)
	if err != nil {
		return nil, err
	}

	return &OggAudio{file: file, reader: ogg, lastGranule: 0, channels: h.Channels, sampleRate: h.SampleRate}, nil
}

func (o *OggAudio) Next() ([]byte, float64, error) {
	pageData, pageHeader, oggErr := o.reader.ParseNextPage()
	if oggErr == io.EOF {
		return nil, 0, io.EOF
	}

	if oggErr != nil {
		return nil, 0, oggErr
	}

	if o.lastGranule == 0 {
		o.lastGranule = pageHeader.GranulePosition
	}

	sampleCount := float64(pageHeader.GranulePosition - o.lastGranule)
	o.lastGranule = pageHeader.GranulePosition

	return pageData, sampleCount, nil
}

func (o *OggAudio) Reset() error {
	o.reader.ResetReader(func(bytesRead int64) io.ReadSeeker {
		return o.file
	})
	return nil
}

func (o OggAudio) Close() error {
	return o.file.Close()
}

func (o OggAudio) GetSampleRate() float64 {
	return float64(o.sampleRate)
}

func (o OggAudio) GetLastGranule() uint64 {
	return o.lastGranule
}

func (o OggAudio) GetSleepDuration(sampleCount float64) time.Duration {
	return time.Duration((sampleCount/o.GetSampleRate())*1000) * time.Millisecond
}
