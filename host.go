package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/kr/pty"
	"github.com/maxmcd/webtty/pkg/sd"
	"github.com/pion/webrtc/v3"
)

type hostSession struct {
	session
	cmd            []string
	nonInteractive bool
	oneWay         bool
	ptmx           *os.File
	ptmxReady      bool
	tmux           bool
}

func logInfo(msg string) {
	l := log.New(os.Stderr, "", 0)
	l.Println(msg)
}

func logError(err error) {
	l := log.New(os.Stderr, "", 0)
	l.Printf("%s\n", err)
}

func (hs *hostSession) dataChannelOnOpen() func() {
	return func() {
		cmd := exec.Command(hs.cmd[0], hs.cmd[1:]...)
		var err error
		hs.ptmx, err = pty.Start(cmd)
		if err != nil {
			logError(err)
			hs.errChan <- err
			return
		}
		hs.ptmxReady = true

		if !hs.nonInteractive {
			if err = hs.makeRawTerminal(); err != nil {
				logError(err)
				hs.errChan <- err
				return
			}
			go func() {
				if _, err = io.Copy(hs.ptmx, os.Stdin); err != nil {
					logError(err)
				}
			}()
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for range c {
				log.Println("Sigint")
				hs.errChan <- errors.New("sigint")
			}
		}()

		buf := make([]byte, 1024)
		for {
			nr, err := hs.ptmx.Read(buf)
			if err != nil {
				if err == io.EOF {
					err = nil
				} else {
					logError(err)
				}
				hs.errChan <- err
				return
			}
			if !hs.nonInteractive {
				if _, err = os.Stdout.Write(buf[0:nr]); err != nil {
					logError(err)
					hs.errChan <- err
					return
				}
			}
			if err = hs.dc.Send(buf[0:nr]); err != nil {
				logError(err)
				hs.errChan <- err
				return
			}
		}
	}
}

func (hs *hostSession) dataChannelOnMessage() func(payload webrtc.DataChannelMessage) {
	return func(p webrtc.DataChannelMessage) {

		// OnMessage can fire before onOpen
		// Let's wait for the pty session to be ready
		for hs.ptmxReady != true {
			time.Sleep(1 * time.Millisecond)
		}

		if p.IsString {
			if len(p.Data) > 2 && p.Data[0] == '[' && p.Data[1] == '"' {
				var msg []string
				err := json.Unmarshal(p.Data, &msg)
				if len(msg) == 0 {
					logError(err)
					hs.errChan <- err
				}
				if msg[0] == "stdin" {
					toWrite := []byte(msg[1])
					if len(toWrite) == 0 {
						return
					}
					_, err := hs.ptmx.Write([]byte(msg[1]))
					if err != nil {
						logError(err)
						hs.errChan <- err
					}
					return
				}
				if msg[0] == "set_size" {
					var size []int
					_ = json.Unmarshal(p.Data, &size)
					ws, err := pty.GetsizeFull(hs.ptmx)
					if err != nil {
						logError(err)
						hs.errChan <- err
						return
					}
					ws.Rows = uint16(size[1])
					ws.Cols = uint16(size[2])

					if len(size) >= 5 {
						ws.X = uint16(size[3])
						ws.Y = uint16(size[4])
					}

					if err := pty.Setsize(hs.ptmx, ws); err != nil {
						logError(err)
						hs.errChan <- err
					}
					return
				}
			}
			if string(p.Data) == "quit" {
				hs.errChan <- nil
				return
			}
			hs.errChan <- fmt.Errorf(
				`Unmatched string message: "%s"`,
				string(p.Data),
			)
		} else {
			_, err := hs.ptmx.Write(p.Data)
			if err != nil {
				logError(err)
				hs.errChan <- err
			}
		}
	}
}

func (hs *hostSession) onDataChannel() func(dc *webrtc.DataChannel) {
	return func(dc *webrtc.DataChannel) {
		hs.dc = dc
		dc.OnOpen(hs.dataChannelOnOpen())
		dc.OnMessage(hs.dataChannelOnMessage())
	}
}

func (hs *hostSession) mustReadStdin() (string, error) {
	var input string
	fmt.Scanln(&input)
	sd, err := sd.Decode(input)
	return sd.Sdp, err
}

func (hs *hostSession) createOffer() (err error) {
	hs.pc.OnDataChannel(hs.onDataChannel())

	// Create unused DataChannel, the offer doesn't implictly have
	// any media sections otherwise
	if _, err = hs.pc.CreateDataChannel("offerer-channel", nil); err != nil {
		logError(err)
		return
	}

	// Create an offer to send to the browser
	offer, err := hs.pc.CreateOffer(nil)
	if err != nil {
		logError(err)
		return
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(hs.pc)

	err = hs.pc.SetLocalDescription(offer)
	if err != nil {
		logError(err)
		return
	}

	// Block until ICE Gathering is complete
	<-gatherComplete

	hs.offer = sd.SessionDescription{
		Sdp: hs.pc.LocalDescription().SDP,
	}
	if hs.oneWay {
		hs.offer.GenKeys()
		hs.offer.Encrypt()
		hs.offer.TenKbSiteLoc = randSeq(100)
	}
	return
}

func (hs *hostSession) run() (err error) {
	if err = hs.init(); err != nil {
		return
	}

	if err = hs.createOffer(); err != nil {
		return
	}
	return
}

func (hs *hostSession) setHostRemoteDescriptionAndWait() (err error) {
	// Set the remote SessionDescription
	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  hs.answer.Sdp,
	}

	// Apply the answer as the remote description
	if err = hs.pc.SetRemoteDescription(answer); err != nil {
		logError(err)
		return
	}

	// Wait to quit
	err = <-hs.errChan
	hs.cleanup()
	return
}
