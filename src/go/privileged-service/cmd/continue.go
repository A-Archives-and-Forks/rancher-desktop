/*
Copyright © 2021 SUSE LLC

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
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"

	"github.com/rancher-sandbox/rancher-desktop/src/go/privileged-service/pkg/manage"
)

// continueCmd represents the continue command
var continueCmd = &cobra.Command{
	Use:   "continue",
	Short: "continue the Rancher Desktop Privileged Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return manage.ControlService(svcName, svc.Continue, svc.Running)
	},
}

func init() {
	rootCmd.AddCommand(continueCmd)
}
