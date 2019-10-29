// Copyright 2019 KIM Keep In Mind GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kim-company/pmux/pwrap"
	"github.com/spf13/cobra"
)

var (
	configPath string
	sockPath   string
)

// mockCmd represents the mockcmd command
var mockCmd = &cobra.Command{
	Use:   "mockcmd",
	Short: "A default mocked command which can be executed by pmux, but does not do anything useful.",
	Run: func(cmd *cobra.Command, args []string) {
		var w io.WriteCloser = os.Stdout
		ctx, cancel := context.WithCancel(context.Background())
		if sockPath != "" {
			br, err := pwrap.NewUnixCommBridge(ctx, sockPath, interProcessCmdHandler(cancel))
			if err != nil {
				log.Printf("[ERROR] %v", err)
				return
			}
			defer br.Close()
			w = br
			go br.Open(ctx)
		}

		for i := 0; ; i++ {
			fmt.Fprintf(w, "waiting %d...", i)
			select {
			case <-time.After(time.Millisecond * 1000):
				fmt.Fprintf(w, "done!\n")
			case <-ctx.Done():
				log.Printf("[INFO] exiting: %v", ctx.Err())
				return
			}
		}
	},
}

func interProcessCmdHandler(cancel context.CancelFunc) func(*pwrap.UnixCommBridge) {
	return pwrap.OnCommand(func(u *pwrap.UnixCommBridge, cmd string) error {
		log.Printf("[INFO] command received: %v", cmd)
		if strings.Contains(cmd, "cancel") {
			cancel()
			return u.Close()
		}
		return nil
	})
}

func init() {
	mockCmd.Flags().StringVarP(&configPath, "config", "", "config.json", "Path to the configuration file.")
	mockCmd.Flags().StringVarP(&sockPath, "socket-path", "", "", "Path to the communication socket address.")
}

func main() {
	mockCmd.Execute()
}
