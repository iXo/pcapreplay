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

// pcap replayer (with gui if needed, with step by step functionality)

package main

import (
	"os"
	"net"

	"github.com/urfave/cli"

	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"

	"kawaiyume.net/pcapreplay/gui"
	"kawaiyume.net/pcapreplay/pcap"
	"kawaiyume.net/pcapreplay/commons"
)

func createGui() {
	commons.MainWin, commons.MainPane = gui.CreateMainWindow("PCAP Replay")

	commons.InterfacesPane = gui.CreateHPanel(commons.MainPane, "Net interfaces", false)
	commons.Interfaces = gui.CreateComboBox(commons.InterfacesPane, true)
	commons.Interfaces.OnSelected(func(*ui.Combobox) {
		intfs, _ := net.Interfaces()

		commons.IntfId = intfs[commons.Interfaces.Selected()].Name
	})

	commons.ReplayPane = gui.CreateVPanel(commons.MainPane, "Replay", true)
	filePane := ui.NewHorizontalBox()
	commons.FileField = ui.NewEntry()
	filePane.Append(commons.FileField, true)
	fileSearchBtn := ui.NewButton("…")
	fileSearchBtn.OnClicked(func(*ui.Button) {
		commons.PcapFile = ui.OpenFile(commons.MainWin)

		commons.FileField.SetText(commons.PcapFile)
		pcap.Infos(commons.PcapFile)
	})
	filePane.Append(fileSearchBtn, false)
	commons.ReplayPane.Append(filePane, false)

	commons.ReplayPane.Append(ui.NewLabel(" "), true)
	commons.Stats1 = gui.CreateLabeledField(commons.ReplayPane, "Avg packet rate :", false, true)
	commons.Stats2 = gui.CreateLabeledField(commons.ReplayPane, "Stats :", false, true)
	commons.StatPBar = ui.NewProgressBar()
	commons.StatPBar.SetValue(-1)
	commons.ReplayPane.Append(commons.StatPBar, true)
	commons.ReplayPane.Append(ui.NewLabel(" "), true)

	commons.ControlsPane = gui.CreateHPanel(commons.MainPane, "Controls", false)
	commons.ControlsPane.Append(ui.NewLabel(" "), true)

	commons.PlayBtn = ui.NewButton("▶")
	commons.FastPlayBtn = ui.NewButton("▶▶")
	commons.StepPlayBtn = ui.NewButton("▮▶")
	commons.StepOnePlayBtn = ui.NewButton("▮▶¹")
	commons.ResetBtn = ui.NewButton("⟲")

	commons.PlayBtn.OnClicked(func(*ui.Button) {
		commons.ReplayFast = false
		gui.DisableControls()
		go pcap.Replay()
	})
	commons.FastPlayBtn.OnClicked(func(*ui.Button) {
		commons.ReplayFast = true
		gui.DisableControls()
		go pcap.Replay()
	})
	commons.StepPlayBtn.OnClicked(func(*ui.Button) {
		commons.ReplayFast = false
		gui.DisableControls()
		go pcap.ReplayStep(commons.StepSpinBox.Value())
	})
	commons.StepOnePlayBtn.OnClicked(func(*ui.Button) {
		commons.ReplayFast = false
		gui.DisableControls()
		go pcap.ReplayStep(1)
	})
	commons.ResetBtn.OnClicked(func(*ui.Button) {
		commons.Stats2.SetText("Resetted")
		go pcap.EndReplay()
	})

	commons.StepSpinBox = ui.NewSpinbox(1, 5000)

	commons.ControlsPane.Append(commons.PlayBtn, false)
	commons.ControlsPane.Append(commons.FastPlayBtn, false)
	commons.ControlsPane.Append(ui.NewLabel("  "), true)
	commons.ControlsPane.Append(commons.StepSpinBox, false)
	commons.ControlsPane.Append(commons.StepPlayBtn, false)
	commons.ControlsPane.Append(commons.StepOnePlayBtn, false)
	commons.ControlsPane.Append(commons.ResetBtn, false)
	commons.ControlsPane.Append(ui.NewLabel(" "), true)

	commons.MainWin.Show()

	go populateGui()
}

func populateGui() {
	intfs, _ := net.Interfaces()
	for _, intf := range intfs {
		commons.Interfaces.Append(intf.Name)
	}

	if commons.PcapFile != "" && commons.FileField != nil {
		commons.FileField.SetText(commons.PcapFile)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "PCAP Replay"
	app.Version = "1.0.0"
	app.Usage = "pcapreplay"
	app.UsageText = "pcapreplay --intf <interface> [--gui] [--fast] --pcap <pcap file>"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "intf",
			Usage:       "system interface id",
			Destination: &commons.IntfId,
		},
		cli.StringFlag{
			Name:        "pcap",
			Usage:       "pcap file to replay",
			Destination: &commons.PcapFile,
		},
		cli.BoolFlag{
			Name:        "fast",
			Usage:       "replay without the real time between each packets",
			Destination: &commons.ReplayFast,
		},
		cli.BoolFlag{
			Name:        "gui",
			Usage:       "start the helper gui",
			Destination: &commons.WithGui,
		},
	}

	app.Action = func(c *cli.Context) error {
		if commons.WithGui {
			ui.Main(createGui)
		} else {
			pcap.Replay()
		}

		return nil
	}

	app.Run(os.Args)
}
