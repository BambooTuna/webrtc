package publish

import (
	"context"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"io"
	"time"
)

func Publisher(ctx context.Context, track *webrtc.TrackLocalStaticSample, audio Audio) error {
	var err error
	for {
		select {
		case <-ctx.Done():
			break
		default:
			data, sampleCount, audioErr := audio.Next()
			if audioErr == io.EOF {
				break
			}

			if audioErr != nil {
				err = audioErr
				break
			}

			trackErr := track.WriteSample(media.Sample{Data: data, Duration: audio.GetSleepDuration(sampleCount)})
			if trackErr != nil {
				err = trackErr
				break
			}

			time.Sleep(audio.GetSleepDuration(sampleCount))
		}
	}

	if err != nil {
		return err
	}

	return nil
}
