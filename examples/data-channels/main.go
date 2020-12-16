package main

import (
	"context"
	"fmt"
	"github.com/pion/webrtc/v3/examples/data-channels/publish"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/examples/internal/signal"
)

func main() {
	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.
	mediaEngine := webrtc.MediaEngine{}
	_ = mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2, SDPFmtpLine: "minptime=10;useinbandfec=1"},
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio)
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTransceiverFromTrack(audioTrack, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly})
	if err != nil {
		panic(err)
	}

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())
	audioStopCtx, audioStop := context.WithCancel(context.Background())
	go func() {
		fmt.Println("再生開始")
		var audio *publish.OggAudio
		audio, err = publish.NewOggAudio("./output.ogg")
		if err != nil {
			fmt.Println(err)
			return
		}

		<-iceConnectedCtx.Done()
		err = publish.Publisher(audioStopCtx, audioTrack, audio)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("再生が正常に終わりました")
	}()

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		} else if connectionState == webrtc.ICEConnectionStateDisconnected {
			audioStop()
		}
	})

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		// Register channel opening handling
		d.OnOpen(func() {
			fmt.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", d.Label(), d.ID())

			for range time.NewTicker(5 * time.Second).C {
				message := signal.RandSeq(15)
				fmt.Printf("Sending '%s'\n", message)

				// Send the message as text
				sendErr := d.SendText(message)
				if sendErr != nil {
					panic(sendErr)
				}
			}
		})

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
		})
	})

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	signal.Decode(signal.MustReadStdin(), &offer)

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	fmt.Println(signal.Encode(*peerConnection.LocalDescription()))

	// Block forever
	select {}
}
