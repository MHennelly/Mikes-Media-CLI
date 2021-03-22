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
	"os"
	"net/http"
	"io"
	"io/ioutil"
	"encoding/json"
	"github.com/spf13/cobra"
	"github.com/joho/godotenv"
	"strings"
	"bytes"
)

// Also used in other commands
type MediaFile struct {
	Name string
	Artist string
	URL string
}

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync media files with backend",
	Long: `Sync media files with backend`,
	Run: func(cmd *cobra.Command, args []string) {
		sync()
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.
	godotenv.Load()

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func sync() {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", os.Getenv("DEST"), nil)
	fmt.Println("Retrieving media data from server...")
	req.Header.Add(os.Getenv("FIELD"), os.Getenv("PASS"))
	fmt.Println("Done")
	resp, _ := client.Do(req)
	defer resp.Body.Close()


	var arr []MediaFile

	b, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(b, &arr)
	var total int64 = 0

	r := bytes.NewReader(b)
	file, _ := os.Create(os.Getenv("STORAGE") + "media.json")
	io.Copy(file, r)
	defer file.Close()

	fmt.Printf("Total # of files: %d\n", len(arr))
	fmt.Println("Syncing...")
	for i:=0; i < len(arr); i++ {
		segments := strings.Split(arr[i].URL, "/")
		_, err := os.Stat(os.Getenv("STORAGE") + segments[len(segments) - 1])
		if err == nil {
			continue
		}
		fmt.Printf("File Name: %s | Artist: %s ...\n", arr[i].Name, arr[i].Artist)
		file, err := os.Create(os.Getenv("STORAGE") + segments[len(segments)-1])
		if err != nil {
			continue
		}
		segments = append(segments, "")
		copy(segments[4:], segments[3:])
		segments[3] = os.Getenv("DEST2")
		req, _ := http.NewRequest("GET", strings.Join(segments, "/"), nil)
		req.Header.Add(os.Getenv("FIELD"), os.Getenv("PASS"))
		resp, _ = client.Do(req)
		defer resp.Body.Close()
		size, _ := io.Copy(file, resp.Body)
		fmt.Printf("Downloaded %.2e bytes \n", float64(size))
		defer file.Close()
		total += size
	}
	fmt.Printf("Sync Complete. Total Bytes Downloaded: %.2e \n", float64(total))
	
}
