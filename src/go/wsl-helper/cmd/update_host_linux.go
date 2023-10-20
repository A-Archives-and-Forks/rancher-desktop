/*
Copyright © 2023 SUSE LLC

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

	"github.com/rancher-sandbox/rancher-desktop/src/go/wsl-helper/pkg/host"
)

var updateHostFileCmd = &cobra.Command{
	Use:   "update-host",
	Short: "Appends a given entry to host file",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := cmd.Flags().GetStringSlice("entries")
		if err != nil {
			return err
		}
		return host.AppendHostFile(entries)
	},
}

func init() {
	updateHostFileCmd.Flags().StringSlice(
		"entries",
		[]string{fmt.Sprintf("%s,%s", host.GatewayIP, host.GatewayDomain)},
		"Array of host file entries to append")
	updateHostFileCmd.MarkFlagRequired("entries")
	rootCmd.AddCommand(updateHostFileCmd)
}
