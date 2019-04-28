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

// gui helpers

package gui

import (
	"github.com/andlabs/ui"

	"pcapreplay/commons"
)

func CreateMainWindow(title string) (*ui.Window, *ui.Box) {
	win := ui.NewWindow(title, 0, 400, false)
	win.SetBorderless(false)
	win.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		win.Destroy()
		return true
	})

	mainPane := ui.NewVerticalBox()
	win.SetChild(mainPane)
	win.SetMargined(true)

	return win, mainPane
}

func CreateHPanel(parent *ui.Box, title string, stretchy bool) *ui.Box {
	tab := ui.NewTab()
	parent.Append(tab, stretchy)

	pane := ui.NewHorizontalBox()
	tab.Append(title, pane)
	tab.SetMargined(0, true)

	return pane
}

func CreateVPanel(parent *ui.Box, title string, stretchy bool) *ui.Box {
	tab := ui.NewTab()
	parent.Append(tab, stretchy)

	pane := ui.NewVerticalBox()
	tab.Append(title, pane)
	tab.SetMargined(0, true)

	return pane
}

func CreateComboBox(parent *ui.Box, stretchy bool) *ui.Combobox {
	cbx := ui.NewCombobox()
	parent.Append(cbx, stretchy)

	return cbx
}

func CreateLabeledField(parent *ui.Box, label string, final bool, readonly bool) *ui.Entry {
	entry := ui.NewEntry()

	parent.Append(ui.NewLabel(label), false)
	parent.Append(entry, false)

	if !final {
		parent.Append(ui.NewLabel(" "), true)
	}

	if readonly {
		entry.SetReadOnly(true)
		entry.Disable()
	}

	return entry
}

func EnableControls() {
	commons.PlayBtn.Enable()
	commons.FastPlayBtn.Enable()
	commons.StepPlayBtn.Enable()
	commons.StepOnePlayBtn.Enable()
	commons.ResetBtn.Enable()

	commons.StepSpinBox.Enable()
}

func DisableControls() {
	commons.PlayBtn.Disable()
	commons.FastPlayBtn.Disable()
	commons.StepPlayBtn.Disable()
	commons.StepOnePlayBtn.Disable()
	commons.ResetBtn.Disable()

	commons.StepSpinBox.Disable()
}
