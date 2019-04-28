// vim: set filetype=go:

/*
BSD 3-Clause License

Copyright (c) 2019, iXo
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

* Neither the name of the copyright holder nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

// pcap helpers

package pcap

import (
	"time"
	"log"
	"io"
	"fmt"

	"github.com/andlabs/ui"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	"pcapreplay/commons"
	"pcapreplay/gui"
)

// replay a packet with a sleep if need after
func WritePacketDelayed(handle *pcap.Handle, buf []byte, ci gopacket.CaptureInfo) {
	if ci.CaptureLength != ci.Length {
		// do not write truncated packets
		return
	}

	intervalInCapture := ci.Timestamp.Sub(commons.LastTS)
	elapsedTime := time.Since(commons.LastSend)

	if (intervalInCapture > elapsedTime) && !commons.LastSend.IsZero() {
		time.Sleep(intervalInCapture - elapsedTime)
	}

	commons.LastSend = time.Now()
	WritePacket(handle, buf)
	commons.LastTS = ci.Timestamp
}

// replay a packet
func WritePacket(handle *pcap.Handle, buf []byte) error {
	if err := handle.WritePacketData(buf); err != nil {
		log.Printf("Failed to send packet: %s\n", err)
		return err
	}
	return nil
}

// get pcap file infos : start time, end time, total number of packets, total size of all packets
func Infos(filename string) (start time.Time, end time.Time, packets int, size int) {
	handleRead, err := pcap.OpenOffline(filename)
	if err != nil {
		log.Fatal("PCAP OpenOffline error (handle to read packet):", err)
	}
	defer handleRead.Close()

	var previousTs time.Time
	var deltaTotal time.Duration

	for {
		data, ci, err := handleRead.ReadPacketData()
		if err != nil && err != io.EOF {
			log.Fatal(err)
		} else if err == io.EOF {
			break
		} else {

			if start.IsZero() {
				start = ci.Timestamp
			}
			end = ci.Timestamp
			packets++
			size += len(data)

			if previousTs.IsZero() {
				previousTs = ci.Timestamp
			} else {
				deltaTotal += ci.Timestamp.Sub(previousTs)
				previousTs = ci.Timestamp
			}
		}
	}
	sec := int(deltaTotal.Seconds())
	if sec == 0 {
		sec = 1
	}

	fmt.Printf("Avg packet rate %d/s\n", packets/sec)
	if commons.WithGui {
		s := fmt.Sprintf("%d", packets/sec)
		commons.Stats1.SetText(s)
	}

	return start, end, packets, size
}

func LoadPcap(filename string) {
	var err error
	commons.PcapHndl, err = pcap.OpenOffline(filename)
	if err != nil {
		log.Fatal("PCAP OpenOffline error (handle to read packet):", err)
	}

	commons.Start = time.Now()
	commons.Pkt = 0
	commons.TsStart, commons.TsEnd, commons.Packets, commons.Size = Infos(commons.PcapFile)
}

func OpenDest(netintf string) *pcap.Handle {
	inactive, crErr := pcap.NewInactiveHandle(netintf)
	if crErr != nil {
		log.Fatal("Error creating inactive handle : %v", crErr)
	}
	inactive.SetPromisc(true)

	defer inactive.CleanUp()

	handleWrite, _ := inactive.Activate()

	return handleWrite
}

func EndReplay() {
	if commons.PcapHndl != nil {
		commons.PcapHndl.Close()
		commons.PcapHndl = nil
	}

	if commons.WithGui {
		ui.QueueMain(func() {
			commons.StatPBar.SetValue(-1)
		})
	}
}

func InternalReplay(handleWrite *pcap.Handle) bool {
	data, ci, err := commons.PcapHndl.ReadPacketData()
	switch {
	case err == io.EOF:
		fmt.Printf("\nFinished in %s", time.Since(commons.Start))
		EndReplay()

		if commons.WithGui {
			ui.QueueMain(func() {
				s := fmt.Sprintf("Finished in %s", time.Since(commons.Start))
				commons.Stats2.SetText(s)

				gui.EnableControls()

				commons.PcapHndl = nil
			})
		}
		return true

	case err != nil:
		log.Printf("Failed to read packet %d: %s\n", commons.Pkt, err)

	default:
		if commons.ReplayFast {
			WritePacket(handleWrite, data)
		} else {
			WritePacketDelayed(handleWrite, data, ci)
		}

		commons.BytesSent += len(data)
		duration := time.Since(commons.Start)
		commons.Pkt++

		if duration > time.Second {
			rate := commons.BytesSent / int(duration.Seconds())
			remainingTime := commons.TsEnd.Sub(commons.TsStart) - duration
			fmt.Printf("\rrate %d kB/sec - sent %d/%d kB - %d/%d packets - remaining time %s", rate/1000, commons.BytesSent/1000, commons.Size/1000, commons.Pkt, commons.Packets, remainingTime)

			if commons.WithGui && !commons.ReplayFast {
				ui.QueueMain(func() {
					commons.StatPBar.SetValue(commons.Pkt * 100 / commons.Packets)
					s := fmt.Sprintf("rate %d kB/sec - sent %d/%d kB - %d/%d packets - remaining time %s", rate/1000, commons.BytesSent/1000, commons.Size/1000, commons.Pkt, commons.Packets, remainingTime)
					commons.Stats2.SetText(s)
				})
			}
		}
	}

	return false
}

// replay a pcap file on the device defined
func Replay() {
	if commons.PcapHndl != nil {
		EndReplay()
	}

	LoadPcap(commons.PcapFile)
	handleWrite := OpenDest(commons.IntfId)

	// loop over packets and write them
	for {
		if InternalReplay(handleWrite) {
			break
		}
	}
}

func ReplayStep(step int) {
	if commons.PcapHndl == nil {
		LoadPcap(commons.PcapFile)
	}

	handleWrite := OpenDest(commons.IntfId)

	// loop over packets and write them
	for i := 0; i < step; i++ {
		if InternalReplay(handleWrite) {
			break
		}
	}

	ui.QueueMain(func() {
		gui.EnableControls()
	})
}
