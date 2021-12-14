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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"

	"github.com/spf13/cobra"
)

var validator1PubKey string
var validator2PubKey string
var validator3PubKey string

func generateValidatorKeys(validatorNumber int64) {
	chainExecutable, _ := exec.LookPath("test-chaind")

	validatorNumberStr := strconv.Itoa((int(validatorNumber)))

	initCmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "init", "validator-" + validatorNumberStr, "-o"},
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

	addValidatorKeyCmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "keys", "add", "validator-" + validatorNumberStr, "--keyring-backend", "test"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := addValidatorKeyCmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	storeValidatorAddressCmd := &exec.Cmd{
		Path: chainExecutable,
		Args: []string{chainExecutable, "keys", "show", "validator-" + validatorNumberStr, "-a", "--keyring-backend", "test"},
		// Stdout: os.Stdout,
		// Stderr: os.Stderr,
	}

	out, err := storeValidatorAddressCmd.CombinedOutput()
	if err != nil {
		fmt.Print("error: ", err)
	}

	fmt.Println("OUTPUT: ", string(out))

	// ok now we want to add the key

	// ok it's now inited
	// we need to do this 3 times

	validatorPubKeyCmd := &exec.Cmd{
		Path: chainExecutable,
		Args: []string{chainExecutable, "tendermint", "show-validator"},
	}

	out, err = validatorPubKeyCmd.CombinedOutput()
	if err != nil {
		fmt.Print("error: ", err)
	}

	// fmt.Println("OUTPUT: ", string(out))

	var jsonMap map[string]string
	json.Unmarshal([]byte(string(out)), &jsonMap)

	fmt.Print(jsonMap["key"])
	if validatorNumberStr == "1" {
		validator1PubKey = string(out)
	} else if validatorNumberStr == "2" {
		validator2PubKey = string(out)
	} else {
		validator3PubKey = string(out)
	}

	e := os.Rename(dir+"/.test-chain/config/node_key.json", dir+"/.test-chain/config/node_key_"+validatorNumberStr+".json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
	e = os.Rename(dir+"/.test-chain/config/priv_validator_key.json", dir+"/.test-chain/config/priv_validator_key_"+validatorNumberStr+".json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
}

func generateBuildArtifacts() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	os.Chdir(dir + "/test-chain")
	// now I can build
	// starportExecutable, _ := exec.LookPath("starport")

	// buildExecutable := &exec.Cmd{
	// 	Path:   starportExecutable,
	// 	Args:   []string{starportExecutable, "chain", "build", "--output", dir + "/test-chain/" + "dist"},
	// 	Stdout: os.Stdout,
	// 	Stderr: os.Stderr,
	// }

	// if err := buildExecutable.Run(); err != nil {
	// 	fmt.Println("error: ", err)
	// }

	// the build should be sitting in /dist now

	// I need to copy over the config folder

	fmt.Println(dir + "/.test-chain")

	fmt.Println(dir + "/test-chain")

	copyConfigFolderCMD := exec.Command("cp", "-r", dir+"/.test-chain", dir+"/test-chain")

	if err := copyConfigFolderCMD.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	// so now it should be test-chain/dist/binary and test-chain/dist/.test-chain
}

// generateTestNetCmd represents the generateTestNet command
var generateTestNetCmd = &cobra.Command{
	Use:   "generate-test-net",
	Short: "One click testnet for starport scaffolded applications",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generateTestNet called")
		// generateValidatorKeys(1)
		// generateValidatorKeys(2)
		// generateValidatorKeys(3)

		// chainExecutable, _ := exec.LookPath("test-chaind")

		// usr, _ := user.Current()
		// dir := usr.HomeDir

		// e := os.Rename(dir+"/.test-chain/config/node_key_1.json", dir+"/.test-chain/config/node_key.json")
		// if e != nil {
		// 	fmt.Println("rename error: ", e)
		// }
		// e = os.Rename(dir+"/.test-chain/config/priv_validator_key_1.json", dir+"/.test-chain/config/priv_validator_key.json")
		// if e != nil {
		// 	fmt.Println("rename error: ", e)
		// }

		// addGenesisAccountValidator1Cmd := &exec.Cmd{
		// 	Path:   chainExecutable,
		// 	Args:   []string{chainExecutable, "add-genesis-account", "validator-1", "100000000000stake", "--keyring-backend", "test"},
		// 	Stdout: os.Stdout,
		// 	Stderr: os.Stderr,
		// }

		// if err := addGenesisAccountValidator1Cmd.Run(); err != nil {
		// 	fmt.Println("error: ", err)
		// }

		// createGentXValidator1Cmd := &exec.Cmd{
		// 	Path:   chainExecutable,
		// 	Args:   []string{chainExecutable, "gentx", "validator-1", "100000000stake", "--chain-id", "test-chain", "--keyring-backend", "test", "--pubkey", validator1PubKey},
		// 	Stdout: os.Stdout,
		// 	Stderr: os.Stderr,
		// }

		// if err := createGentXValidator1Cmd.Run(); err != nil {
		// 	fmt.Println("error: ", err)
		// }
		// e = os.Rename(dir+"/.test-chain/config/node_key.json", dir+"/.test-chain/config/node_key_1.json")
		// if e != nil {
		// 	fmt.Println("rename error: ", e)
		// }
		// e = os.Rename(dir+"/.test-chain/config/priv_validator_key.json", dir+"/.test-chain/config/priv_validator_key_1.json")
		// if e != nil {
		// 	fmt.Println("rename error: ", e)
		// }

		// e = os.Rename(dir+"/.test-chain/config/node_key_2.json", dir+"/.test-chain/config/node_key.json")
		// if e != nil {
		// 	fmt.Println("rename error: ", e)
		// }
		// e = os.Rename(dir+"/.test-chain/config/priv_validator_key_2.json", dir+"/.test-chain/config/priv_validator_key.json")
		// if e != nil {
		// 	fmt.Println("rename error: ", e)
		// }

		// fmt.Println(("RUNNIG GENESIS ACCOUNT FOR VALIDATOR 2"))

		// addGenesisAccountValidator2Cmd := &exec.Cmd{
		// 	Path:   chainExecutable,
		// 	Args:   []string{chainExecutable, "add-genesis-account", "validator-2", "100000000000stake", "--keyring-backend", "test"},
		// 	Stdout: os.Stdout,
		// 	Stderr: os.Stderr,
		// }

		// if err := addGenesisAccountValidator2Cmd.Run(); err != nil {
		// 	fmt.Println("error: ", err)
		// }

		// createGentXValidator2Cmd := &exec.Cmd{
		// 	Path:   chainExecutable,
		// 	Args:   []string{chainExecutable, "gentx", "validator-2", "100000000stake", "--chain-id", "test-chain", "--keyring-backend", "test", "--pubkey", validator2PubKey},
		// 	Stdout: os.Stdout,
		// 	Stderr: os.Stderr,
		// }

		// if err := createGentXValidator2Cmd.Run(); err != nil {
		// 	fmt.Println("error: ", err)
		// }

		// e = os.Rename(dir+"/.test-chain/config/node_key.json", dir+"/.test-chain/config/node_key_2.json")
		// if e != nil {
		// 	fmt.Println("rename error: ", e)
		// }
		// e = os.Rename(dir+"/.test-chain/config/priv_validator_key.json", dir+"/.test-chain/config/priv_validator_key_2.json")
		// if e != nil {
		// 	fmt.Println("rename error: ", e)
		// }

		// e = os.Rename(dir+"/.test-chain/config/node_key_3.json", dir+"/.test-chain/config/node_key.json")
		// if e != nil {
		// 	fmt.Println("rename error: ", e)
		// }
		// e = os.Rename(dir+"/.test-chain/config/priv_validator_key_3.json", dir+"/.test-chain/config/priv_validator_key.json")
		// if e != nil {
		// 	fmt.Println("rename error: ", e)
		// }

		// addGenesisAccountValidator3Cmd := &exec.Cmd{
		// 	Path:   chainExecutable,
		// 	Args:   []string{chainExecutable, "add-genesis-account", "validator-3", "100000000000stake", "--keyring-backend", "test"},
		// 	Stdout: os.Stdout,
		// 	Stderr: os.Stderr,
		// }

		// if err := addGenesisAccountValidator3Cmd.Run(); err != nil {
		// 	fmt.Println("error: ", err)
		// }

		// createGentXValidator3Cmd := &exec.Cmd{
		// 	Path:   chainExecutable,
		// 	Args:   []string{chainExecutable, "gentx", "validator-3", "100000000stake", "--chain-id", "test-chain", "--keyring-backend", "test", "--pubkey", validator3PubKey},
		// 	Stdout: os.Stdout,
		// 	Stderr: os.Stderr,
		// }

		// if err := createGentXValidator3Cmd.Run(); err != nil {
		// 	fmt.Println("error: ", err)
		// }

		// collectGentXCmd := &exec.Cmd{
		// 	Path: chainExecutable,
		// 	Args: []string{chainExecutable, "collect-gentxs"},
		// }

		// if err := collectGentXCmd.Run(); err != nil {
		// 	fmt.Println("error: ", err)
		// }

		generateBuildArtifacts()

		// now we want to generate the gentxs?

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
