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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"

	"github.com/spf13/cobra"
)

func generateValidatorKeys(validatorNumber int64) {
	chainExecutable, _ := exec.LookPath("test-chaind")

	initCmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "init", "validator-" + strconv.Itoa(int(validatorNumber)), "-o"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := initCmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	usr, _ := user.Current()
	dir := usr.HomeDir

	fmt.Println(dir)
	if _, error := os.Stat(dir + "/.test-chain/config"); error == nil {
		fmt.Println("it exists!")
	} else if errors.Is(error, os.ErrNotExist) {
		fmt.Println("it does not exist!")
	}

	// ok it's now inited
	// we need to do this 3 times
	e := os.Rename(dir+"/.test-chain/config/node_key.json", dir+"/.test-chain/config/node_key_"+strconv.Itoa(int(validatorNumber))+".json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
	e = os.Rename(dir+"/.test-chain/config/priv_validator_key.json", dir+"/.test-chain/config/priv_validator_key_"+strconv.Itoa(int(validatorNumber))+".json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}

}

// generateTestNetCmd represents the generateTestNet command
var generateTestNetCmd = &cobra.Command{
	Use:   "generate-test-net",
	Short: "One click testnet for starport scaffolded applications",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generateTestNet called")
		generateValidatorKeys(1)
		generateValidatorKeys(2)
		generateValidatorKeys(3)

		// so I am in the folder right now, the chain is scaffolded, the executable exists in golang.

		// I want to call init.
		// TODO: Set up the infrastructure
		/*
			1. Chain is scaffolded
			2.

		*/

	},
}

func init() {
	rootCmd.AddCommand(generateTestNetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateTestNetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateTestNetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
