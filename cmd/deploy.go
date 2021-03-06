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
	"os/user"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/prakashsanker/one-click-cosmos-testnet/config"
	"github.com/prakashsanker/one-click-cosmos-testnet/testnet"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your chain to your validators",
	Long:  `Run testnet deploy <SHA> in order to deploy a specific version of your code. Default is HEAD`,
	Run: func(cmd *cobra.Command, args []string) {
		rmExecutable, _ := exec.LookPath("rm")
		usr, _ := user.Current()
		dir := usr.HomeDir
		commitSha, _ := cmd.Flags().GetString("sha")

		rmConfigFolder := &exec.Cmd{
			Path:   rmExecutable,
			Args:   []string{rmExecutable, "-rf", dir + "/" + config.GetChainConfigFolderName()},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := rmConfigFolder.Run(); err != nil {
			fmt.Println("deploy error: ", err)
		}
		if commitSha != "" {
			fmt.Println("Commit sha provided... using " + commitSha)
			r, err := git.PlainOpen(dir + "/" + config.GetChainFolderName())
			testnet.CheckIfError(err)

			worktree, err := r.Worktree()

			if err != nil {
				fmt.Println(err)
			}

			err = worktree.Checkout(&git.CheckoutOptions{
				Hash: plumbing.NewHash(commitSha),
			})
			testnet.CheckIfError(err)

			testnet.Setup()
			testnet.GenerateValidatorKeys(1)
			testnet.GenerateValidatorKeys(2)
			testnet.GenerateValidatorKeys(3)

			testnet.GenerateGenesisTransactionsAndAccounts()

			testnet.GenerateBuildArtifacts(commitSha)
			testnet.PushToEcr(commitSha)

			testnet.ConfigureValidators()
			testnet.UpdateValidators()
			testnet.GetSummaryInformation()
			// we need to also store the branch we were on before.

		} else {
			fmt.Println("No commit sha provided....using current HEAD")
			latestSha := testnet.GetLatestSha()
			testnet.Setup()
			testnet.GenerateValidatorKeys(1)
			testnet.GenerateValidatorKeys(2)
			testnet.GenerateValidatorKeys(3)

			testnet.GenerateGenesisTransactionsAndAccounts()

			testnet.GenerateBuildArtifacts(latestSha)
			testnet.PushToEcr(latestSha)

			testnet.ConfigureValidators()
			testnet.UpdateValidators()
			testnet.GetSummaryInformation()
		}

	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	rootCmd.PersistentFlags().String("sha", "", "A github commit to deploy to your chain")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
