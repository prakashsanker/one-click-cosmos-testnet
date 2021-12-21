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
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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
	dockerExecutable, _ := exec.LookPath("docker")
	os.Chdir(dir + "/test-chain")

	fmt.Println(dir + "/.test-chain")

	fmt.Println(dir + "/test-chain")

	copyConfigFolderCMD := exec.Command("cp", "-r", dir+"/.test-chain", dir+"/test-chain")

	if err := copyConfigFolderCMD.Run(); err != nil {
		fmt.Println("error: ", err)
	}
	// so now it should be test-chain/dist/binary and test-chain/dist/.test-chain

	// now we want to build the Docker image

	buildDockerImage := &exec.Cmd{
		Path:   dockerExecutable,
		Args:   []string{dockerExecutable, "buildx", "build", "--platform", "linux/amd64", "-f", dir + "/one-click-cosmos-testnet/Dockerfile", dir + "/test-chain", "-t", "test-chain", "--no-cache"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := buildDockerImage.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	tagDockerImage := &exec.Cmd{
		Path:   dockerExecutable,
		Args:   []string{dockerExecutable, "tag", "test-chain:latest", "187926495729.dkr.ecr.ap-south-1.amazonaws.com/one-click-cosmos-testnet-repo:latest"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := tagDockerImage.Run(); err != nil {
		fmt.Println("error: ", err)
	}
}

func updateValidators() {
	awsExecutable, _ := exec.LookPath("aws")
	updateECSServiceCMD := &exec.Cmd{
		Path:   awsExecutable,
		Args:   []string{awsExecutable, "ecs", "update-service", "--cluster", "testnet-cluster", "--service", "testnet-app", "--force-new-deployment"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := updateECSServiceCMD.Run(); err != nil {
		fmt.Println("error: ", err)
	}
}

type EC2Instance struct {
	DnsName    string
	LaunchTime time.Time
}

func moveConfigIntoValidatorConfigFolder(dnsName string, validatorNumber int) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	suffix := ""
	if validatorNumber > 0 {
		suffix = "_" + strconv.Itoa(validatorNumber)
	}

	fmt.Println(dir + "/.test-chain/config/validator-config")

	err := os.Mkdir(dir+"/.test-chain/config/validator-config", 0770)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(dir + "/.test-chain/config/node_key" + suffix + ".json")
	fmt.Println(dir + "/.test-chain/config/validator-config/")
	cpExecutable, _ := exec.LookPath("cp")

	moveNodeKey := &exec.Cmd{
		Path:   cpExecutable,
		Args:   []string{cpExecutable, dir + "/.test-chain/config/node_key" + suffix + ".json", dir + "/.test-chain/config/validator-config/node_key.json"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := moveNodeKey.Run(); err != nil {
		fmt.Println("Move Node Error: ", err)
	}

	moveValidatorKey := &exec.Cmd{
		Path:   cpExecutable,
		Args:   []string{cpExecutable, dir + "/.test-chain/config/priv_validator_key" + suffix + ".json", dir + "/.test-chain/config/validator-config/priv_validator_key.json"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := moveValidatorKey.Run(); err != nil {
		fmt.Println("Move Validator Error: ", err)
	}
	scpExecutable, _ := exec.LookPath("scp")

	fmt.Println("RUNNING SCP?")

	copyConfig := &exec.Cmd{
		Path:   scpExecutable,
		Args:   []string{scpExecutable, "-i", "validator-key.pem", "-pr", dir + "/.test-chain/config/validator-config", "ec2-user@" + dnsName + ":~/validator-config"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	fmt.Println("are we getting here?")

	if err := copyConfig.Run(); err != nil {
		fmt.Println("Scp error: ", err)
	}

	// rmExecutable, _ := exec.LookPath("rm")

	// rmValidatorConfig := &exec.Cmd{
	// 	Path:   rmExecutable,
	// 	Args:   []string{rmExecutable, "-rf", dir + "/.test-chain/validator-config"},
	// 	Stdout: os.Stdout,
	// 	Stderr: os.Stderr,
	// }

	// if err := rmValidatorConfig.Run(); err != nil {
	// 	fmt.Println("Rm Validator Config error: ", err)
	// }

}

func configureValidators() {
	// Load session from shared config
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	var instances []EC2Instance

	// Create new EC2 client
	ec2Svc := ec2.New(sess)

	// Call to get detailed information on each instance

	result, err := ec2Svc.DescribeInstances(nil)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		fmt.Println("Success", result)
		for _, reservation := range result.Reservations {
			for _, instance := range reservation.Instances {
				fmt.Println(*instance.NetworkInterfaces[0].Association.PublicDnsName)
				publicDnsName := *instance.NetworkInterfaces[0].Association.PublicDnsName
				newInstance := EC2Instance{
					DnsName:    publicDnsName,
					LaunchTime: *instance.LaunchTime,
				}
				instances = append(instances, newInstance)
			}
		}

		sort.Slice(instances, func(i, j int) bool {
			return instances[i].LaunchTime.Before(instances[j].LaunchTime)
		})
		// now I have this I want to

		// 1. Copy over the node_key.json and the priv_validator_key.json --> make sure that they work with the volume mount
		// 2. Then I need to modify the config.toml so that the persistent_peers are updated properly.
		// scpExecutable, _ := exec.LookPath("scp")
		// usr, _ := user.Current()
		// dir := usr.HomeDir

		for i, instance := range instances {
			dnsName := instance.DnsName
			fmt.Println("dns name")
			fmt.Println(dnsName)
			moveConfigIntoValidatorConfigFolder(dnsName, i)
		}

	}
}

func pushToECR() {
	awsExecutable, _ := exec.LookPath("aws")
	dockerExecutable, _ := exec.LookPath("docker")

	ecrGetCredentialsCMD := &exec.Cmd{
		Path: awsExecutable,
		Args: []string{awsExecutable, "ecr", "get-login-password", "--region", "ap-south-1"},
	}

	// if err := ecrGetCredentialsCMD.Run(); err != nil {
	// 	fmt.Println("error: ", err)
	// }
	fmt.Println("ECR GET CREDENTIALS OUTPUT")
	out, err := ecrGetCredentialsCMD.CombinedOutput()
	if err != nil {
		fmt.Print("error: ", err)
	}

	dockerEcrLoginCMD := &exec.Cmd{
		Path:   dockerExecutable,
		Args:   []string{dockerExecutable, "login", "--username", "AWS", "-p", string(out), "187926495729.dkr.ecr.ap-south-1.amazonaws.com"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := dockerEcrLoginCMD.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	dockerPushECRCMD := &exec.Cmd{
		Path:   dockerExecutable,
		Args:   []string{dockerExecutable, "push", "187926495729.dkr.ecr.ap-south-1.amazonaws.com/one-click-cosmos-testnet-repo:latest"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := dockerPushECRCMD.Run(); err != nil {
		fmt.Println("error: ", err)

	}
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

		// generateBuildArtifacts()
		// pushToECR()

		// updateValidators()
		configureValidators()
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
