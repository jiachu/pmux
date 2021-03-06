// SPDX-FileCopyrightText: 2019 KIM KeepInMind GmbH
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pw, close := makeProgressWriter(ctx, cancel, sockPath)
		defer close()

		for i := 0; ; i++ {
			select {
			case <-time.After(time.Millisecond * 1000):
				if err := pw("waited 1 second", -1, -1, i, -1); err != nil {
					log.Printf("[ERROR] %v", err)
				}
			case <-ctx.Done():
				log.Printf("[INFO] exiting: %v", ctx.Err())
				return
			}
		}
	},
}

func writeProgressUpdateDefault(d string, stage, stages, partial, tot int) error {
	fmt.Fprintf(os.Stdout, "%d: %s\n", partial, d)
	return nil
}

func makeProgressWriter(ctx context.Context, cancel context.CancelFunc, sockPath string) (pwrap.WriteProgressUpdateFunc, func()) {
	if sockPath == "" {
		return writeProgressUpdateDefault, func() {}
	}

	br, err := pwrap.NewUnixCommBridge(ctx, sockPath, makeOnCommandOption(cancel))
	if err != nil {
		log.Printf("[ERROR] unable to make progress writer: %v", err)
		return writeProgressUpdateDefault, func() {}
	}
	go br.Open(ctx)
	return br.WriteProgressUpdate, func() {
		br.Close()
	}
}

func makeOnCommandOption(cancel context.CancelFunc) func(*pwrap.UnixCommBridge) {
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
