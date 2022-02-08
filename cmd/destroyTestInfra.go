/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

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
	"os/exec"

	"github.com/spf13/cobra"
)

// destroyTestInfraCmd represents the destroyTestInfra command
var destroyTestInfraCmd = &cobra.Command{
	Use:   "destroy-test-infra",
	Short: "Destroy your validators",
	Long:  `Takes down validators`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("destroyTestInfra called")
		terraformExecutable, _ := exec.LookPath("terraform")
		terraformDestroyCmd := &exec.Cmd{
			Path:   terraformExecutable,
			Args:   []string{terraformExecutable, "destroy", "-auto-approve"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := terraformDestroyCmd.Run(); err != nil {
			fmt.Println("error: ", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyTestInfraCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destroyTestInfraCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyTestInfraCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
