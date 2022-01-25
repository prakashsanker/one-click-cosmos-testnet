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

var generateTestInfra = &cobra.Command{
	Use:   "generate-test-infra",
	Short: "Create cloud validator set",
	Long:  `Run this command to generate the infrastructure for your validators in AWS`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generate Test Infra called")
		terraformExecutable, _ := exec.LookPath("terraform")
		terraformApplyCmd := &exec.Cmd{
			Path:   terraformExecutable,
			Args:   []string{terraformExecutable, "apply", "-auto-approve"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := terraformApplyCmd.Run(); err != nil {
			fmt.Println("error: ", err)
		}

		chmodExecutable, _ := exec.LookPath("chmod")
		chmodValidatorKeyPem := &exec.Cmd{
			Path:   chmodExecutable,
			Args:   []string{chmodExecutable, "400", "validator_key.pem"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := chmodValidatorKeyPem.Run(); err != nil {
			fmt.Println("chmod validator error: ", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateTestInfra)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateTestInfraCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateTestInfraCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}