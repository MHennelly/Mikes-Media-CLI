/*
Copyright © 2021 Michael Hennelly <mike@mhennelly.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"encoding/json"
	"io"
	"io/ioutil"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"strings"
	"math/rand"
	"time"
)

// playCmd represents the play command
var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Launch TUI for playing media files",
	Long: `Launch TUI for playing media files`,
	Run: func(cmd *cobra.Command, args []string) {
		play()
	},
}

func init() {
	rootCmd.AddCommand(playCmd)

	// Here you will define your flags and configuration settings.
	//playCmd.Flags().IntVarP(&ID, "id", "i", 0, "Media file id")
	//playCmd.MarkFlagRequired("id")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// playCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// playCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}



func play() {
	
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	var arr []MediaFile

	data, _ := ioutil.ReadFile(os.Getenv("STORAGE") + "media.json")
	json.Unmarshal(data, &arr)

	l := widgets.NewList()
	l.Title = "Mikes Files"
	l.Rows = []string{}
	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	l.SetRect(0, 0, 100, 100)

	p := widgets.NewParagraph()
	p.Title = "Current Media File"
	p.BorderStyle.Fg = ui.ColorBlue
	p.SetRect(0, 100, 100, 10)

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(0.9, l),
		ui.NewRow(0.1, p),
	)

	ui.Render(grid)

	for i := 0; i < len(arr); i++ {
		l.Rows = append(l.Rows, fmt.Sprintf("[%d] %s - %s", i, arr[i].Name, arr[i].Artist))
	}

	ID := 0
	var stdin io.WriteCloser
	var cmd *exec.Cmd

	rand.Seed(time.Now().UnixNano())

	ui.Render(l, p)

	previousKey := ""
	uiEvents := ui.PollEvents()
	
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			if cmd != nil {
				cmd.Process.Kill()
			}
			return
		case "j", "<Down>":
			l.ScrollDown()
			if ID < len(arr) - 1 {
				ID += 1
			}
		case "k", "<Up>":
			l.ScrollUp()
			if ID > 0 {
				ID -= 1
			}
		case "<Home>":
			l.ScrollTop()
			ID = 0
		case "G", "<End>":
			l.ScrollBottom()
			ID = len(arr) - 1
		case "<Enter>":
			if cmd != nil {
				stdin.Close()
				cmd.Process.Kill()
			}
			segments := strings.Split(arr[ID].URL, "/")
			path := os.Getenv("STORAGE") + segments[len(segments) - 1]
			//shuffle_rest_cmd := ";mpg123 -Z " + os.Getenv("STORAGE") + "*.mp3"
			cmd = exec.Command("mpg123", path)//, shuffle_rest_cmd)
			stdin, _ = cmd.StdinPipe()
			p.Text = fmt.Sprintf("Playing [%d] | %s - %s ...", ID, arr[ID].Name, arr[ID].Artist)
			go cmd.Run()
		case "<Resize>":
			payload := e.Payload.(ui.Resize)
			grid.SetRect(0, 0, payload.Width, payload.Height)
			ui.Clear()
			ui.Render(grid)
		case "f":
			if stdin != nil {
				io.WriteString(stdin, "f\n")
			}
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}
		
		ui.Render(l,p)
	}
}
