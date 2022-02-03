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
		// check to see if the clis exist

		terraformExecutable, terraformErr := exec.LookPath("terraform")
		if terraformErr != nil {
			fmt.Println("Please install terraform")
			return
		}

		_, awsErr := exec.LookPath("aws")
		if awsErr != nil {
			fmt.Println("Please install aws-cli and run one-click-cosmos-testnet configure")
			return
		}

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
}
