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

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Provide configuration to the CLI to generate the testnet",
	Long:  `This is currently your AWS credentials - we will change this to allow for more options soon!`,
	Run: func(cmd *cobra.Command, args []string) {
		accessKeyId, _ := cmd.Flags().GetString("aws-access-key-id")
		secretAccessKey, _ := cmd.Flags().GetString("aws-secret-access-key")
		region, _ := cmd.Flags().GetString("region")

		awsExecutable, _ := exec.LookPath("aws")

		awsConfigureAccessKey := &exec.Cmd{
			Path:   awsExecutable,
			Args:   []string{awsExecutable, "configure", "set", "aws_access_key_id", accessKeyId},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := awsConfigureAccessKey.Run(); err != nil {
			fmt.Println("error: ", err)
		}

		awsConfigureSecretAccessKey := &exec.Cmd{
			Path:   awsExecutable,
			Args:   []string{awsExecutable, "configure", "set", "aws_secret_access_key", secretAccessKey},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := awsConfigureSecretAccessKey.Run(); err != nil {
			fmt.Println("error: ", err)
		}

		awsConfigureRegion := &exec.Cmd{
			Path:   awsExecutable,
			Args:   []string{awsExecutable, "configure", "set", "region", region},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := awsConfigureRegion.Run(); err != nil {
			fmt.Println("error: ", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(configureCmd)

	configureCmd.Flags().String("aws-access-key-id", "", "Your AWS access key id")
	configureCmd.Flags().String("aws-secret-access-key", "", "Your AWS secret access key")
	configureCmd.Flags().String("region", "", "Your AWS region")

	configureCmd.MarkFlagRequired("aws-access-key-id")
	configureCmd.MarkFlagRequired("aws-secret-access-key")
	configureCmd.MarkFlagRequired("region")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
