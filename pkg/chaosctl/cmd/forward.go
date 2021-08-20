// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
)

// completionCmd represents the completion command
var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Forward ctrl api port to local",
	RunE: func(cmd *cobra.Command, args []string) error {
		cancel, port, err := common.ForwardSvcPorts(context.Background(), "chaos-testing", "svc/chaos-mesh-controller-manager", 10082)
		if err != nil {
			return err
		}
		fmt.Printf("forward ctrl api to local port(%d)", port)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		cancel()
		return nil
	},
}
