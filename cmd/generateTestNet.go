/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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

	"github.com/prakashsanker/one-click-cosmos-testnet/testnet"
	"github.com/spf13/cobra"
)

// generateTestNetCmd represents the generateTestNet command
var generateTestNetCmd = &cobra.Command{
	Use:   "generate-test-net",
	Short: "create 3 node validator set",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generateTestNet called")

		/*
			Assumptions -
				1. starport chain init has been run once, so .test-chain exists and a local binary exists at the gopath.
				2. Terraform exists
				3. AWS is configured
		*/

		// We first need to clear up any existing node_key.json and priv_validator_key.json
		testnet.Setup()
		testnet.GenerateValidatorKeys(1)
		testnet.GenerateValidatorKeys(2)
		testnet.GenerateValidatorKeys(3)

		testnet.GenerateGenesisTransactionsAndAccounts()

		testnet.GenerateBuildArtifacts("")
		testnet.PushToEcr("")

		testnet.ConfigureValidators()
		testnet.UpdateValidators()

	},
}

func init() {
	rootCmd.AddCommand(generateTestNetCmd)
	// rootCmd.AddCommand((generateTestInfra))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateTestNetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateTestNetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
