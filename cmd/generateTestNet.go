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
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

var validator1PubKey string
var validator2PubKey string
var validator3PubKey string

var nodeIdsArray [3]string = [3]string{"ec2-13-232-211-73.ap-south-1.compute.amazonaws.com", "ec2-13-234-59-187.ap-south-1.compute.amazonaws.com", "ec2-13-127-55-76.ap-south-1.compute.amazonaws.com"}

func generateValidatorKeys(validatorNumber int64) {

	fmt.Println("calling generate validator keys")
	chainExecutable, _ := exec.LookPath("test-chaind")

	validatorNumberStr := strconv.Itoa((int(validatorNumber)))

	usr, _ := user.Current()
	dir := usr.HomeDir

	addValidatorKeyCmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "keys", "add", "validator-" + validatorNumberStr, "--keyring-backend", "test"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	fmt.Println("executing add validator key command")
	if err := addValidatorKeyCmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	storeValidatorAddressCmd := &exec.Cmd{
		Path: chainExecutable,
		Args: []string{chainExecutable, "keys", "show", "validator-" + validatorNumberStr, "-a", "--keyring-backend", "test"},
	}
	fmt.Println("storing validator address in variable")
	out, err := storeValidatorAddressCmd.CombinedOutput()
	if err != nil {
		fmt.Print("error: ", err)
	}

	// ok now we want to add the key

	// ok it's now inited
	// we need to do this 3 times

	// validatorPubKeyCmd := &exec.Cmd{
	// 	Path: chainExecutable,
	// 	Args: []string{chainExecutable, "tendermint", "show-validator"},
	// }

	// out, err := validatorPubKeyCmd.CombinedOutput()
	// if err != nil {
	// 	fmt.Print("error: ", err)
	// }

	var jsonMap map[string]string
	json.Unmarshal([]byte(string(out)), &jsonMap)
	fmt.Println("storing validator address")
	fmt.Println(validatorNumberStr)
	if validatorNumberStr == "1" {
		fmt.Println("hitting the right if")
		validator1PubKey = string(out)
		fmt.Println("validatorPubKey: ", validator1PubKey)
	} else if validatorNumberStr == "2" {
		validator2PubKey = string(out)
	} else {
		validator3PubKey = string(out)
	}

	// nodeIdCmd := &exec.Cmd{
	// 	Path: chainExecutable,
	// 	Args: []string{chainExecutable, "tendermint", "show-node-id"},
	// }

	// out, err = nodeIdCmd.CombinedOutput()

	if err != nil {
		fmt.Println("error: ", err)
	}

	initCmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "init", "validator-" + validatorNumberStr, "--chain-id", "test-chain"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	fmt.Println("initialize validator")

	if err := initCmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	// nodeIdsArray = append(nodeIdsArray, string(out))
	fmt.Println("renaming node_key.json to node_key_1.json")
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

	rmDistCMD := exec.Command("rm", "-rf", dir+"/test-chain/dist")

	fmt.Println("removing dist folder")

	if err := rmDistCMD.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	mkDistFolderCMD := exec.Command("mkdir", dir+"/test-chain/dist")
	if err := mkDistFolderCMD.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	fmt.Println("copying over test-chain config folder")

	copyConfigFolderCMD := exec.Command("cp", "-R", dir+"/.test-chain", dir+"/test-chain/dist/")

	if err := copyConfigFolderCMD.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	fmt.Println("generating binary")

	generateBinary := exec.Command("starport", "chain", "build", "-o", dir+"/test-chain/dist", "--release", "-t", "linux:amd64")
	// // so now it should be test-chain/dist/binary and test-chain/dist/.test-chain

	if err := generateBinary.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	// // need to untar
	// // TODO: generalize this

	fmt.Println("untarring")
	untarCMD := exec.Command("tar", "-xf", dir+"/test-chain/dist/test-chain_linux_amd64.tar.gz", "-C", dir+"/test-chain/dist/")
	if err := untarCMD.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	// // // now we want to build the Docker image

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
		Args:   []string{scpExecutable, "-i", "./validator_key.pem", "-pr", dir + "/.test-chain/config/validator-config", "ec2-user@" + dnsName + ":~/validator-config"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	fmt.Println("are we getting here?")

	if err := copyConfig.Run(); err != nil {
		fmt.Println("Scp error: ", err)
	}

	copyConfigTomlCmd := &exec.Cmd{
		Path:   scpExecutable,
		Args:   []string{scpExecutable, "-i", "./validator_key.pem", "-pr", dir + "/.test-chain/config/config.toml", "ec2-user@" + dnsName + ":~/validator-config"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := copyConfigTomlCmd.Run(); err != nil {
		fmt.Print("Config.toml error: ", err)
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

func getEC2Instances() []EC2Instance {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	var instances []EC2Instance

	// Create new EC2 client
	ec2Svc := ec2.New(sess)
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
	}
	return instances

}

func constructPersistentPeerString(instances []EC2Instance) string {
	fmt.Println("INSTANCES")
	fmt.Println(instances)

	// nodeIdsArray[0] = "ec2-3-108-42-87.ap-south-1.compute.amazonaws.com"
	// nodeIdsArray[1] = "ec2-52-66-87-194.ap-south-1.compute.amazonaws.com"
	// nodeIdsArray[2] = "ec2-13-233-253-6.ap-south-1.compute.amazonaws.com"

	fmt.Println("NODE IDS ARRAY")
	fmt.Println(nodeIdsArray)
	var persistentPeerString = "persistent_peers = \""
	for i, instance := range instances {
		dnsName := instance.DnsName
		nodeId := nodeIdsArray[i]
		toAdd := nodeId + "@" + dnsName + ":26656,"
		persistentPeerString = persistentPeerString + toAdd
	}

	persistentPeerString += "\""

	var treatedPersistentPeerString = strings.Replace(persistentPeerString, "\n", "", -1)

	return treatedPersistentPeerString
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

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("running")},
			},
		},
	}

	result, err := ec2Svc.DescribeInstances(params)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		for _, reservation := range result.Reservations {
			for _, instance := range reservation.Instances {
				fmt.Println("INSTANCE")
				fmt.Println(instance)
				publicDnsName := *instance.NetworkInterfaces[0].Association.PublicDnsName
				newInstance := EC2Instance{
					DnsName:    publicDnsName,
					LaunchTime: *instance.LaunchTime,
				}
				instances = append(instances, newInstance)
			}
		}

		fmt.Println("instances")

		sort.Slice(instances, func(i, j int) bool {
			return instances[i].LaunchTime.Before(instances[j].LaunchTime)
		})
		// now I have this I want to

		// 1. Copy over the node_key.json and the priv_validator_key.json --> make sure that they work with the volume mount
		// 2. Then I need to modify the config.toml so that the persistent_peers are updated properly.
		// scpExecutable, _ := exec.LookPath("scp")
		// usr, _ := user.Current()
		// dir := usr.HomeDir

		persistentPeerString := constructPersistentPeerString(instances)
		usr, _ := user.Current()
		dir := usr.HomeDir

		// now take the config.toml

		sedExecutable, _ := exec.LookPath("sed")
		addPersistentPeersToConfigCmd := &exec.Cmd{
			Path:   sedExecutable,
			Args:   []string{sedExecutable, "-i", "''", "s/persistent_peers = \"\"/" + persistentPeerString + "/g", dir + "/.test-chain/config/config.toml"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := addPersistentPeersToConfigCmd.Run(); err != nil {
			fmt.Println("error: ", err)
		}

		for i, instance := range instances {
			dnsName := instance.DnsName
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

func setup() {
	// clean up
	usr, _ := user.Current()
	dir := usr.HomeDir

	rmNodeKeyCMD := exec.Command("rm", dir+"/.test-chain/config/node_key.json")

	if err := rmNodeKeyCMD.Run(); err != nil {
		fmt.Println("error: ", err)

	}

	rmValidatorKeyCMD := exec.Command("rm", dir+"/.test-chain/config/priv_validator_key.json")

	if err := rmValidatorKeyCMD.Run(); err != nil {
		fmt.Println("error: ", err)

	}

	rmGenesisJsonCMD := exec.Command("rm", dir+"/.test-chain/config/genesis.json")
	if err := rmGenesisJsonCMD.Run(); err != nil {
		fmt.Println("error: ", err)

	}

	// 	rmDistCMD := exec.Command("rm", "-rf", dir+"/test-chain/dist")

}

func collectGentX() {
	chainExecutable, _ := exec.LookPath("test-chaind")
	collectGentXCmd := &exec.Cmd{
		Path: chainExecutable,
		Args: []string{chainExecutable, "collect-gentxs"},
	}
	fmt.Println("collecting gentxs")

	if err := collectGentXCmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}
}

func generateGenesisTransactionsAndAccounts() {
	chainExecutable, _ := exec.LookPath("test-chaind")
	fmt.Println("generating genesis transactions and accounts")

	usr, _ := user.Current()
	dir := usr.HomeDir
	fmt.Println("renaming node_key_1.json to node_key.json")
	e := os.Rename(dir+"/.test-chain/config/node_key_1.json", dir+"/.test-chain/config/node_key.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
	e = os.Rename(dir+"/.test-chain/config/priv_validator_key_1.json", dir+"/.test-chain/config/priv_validator_key.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}

	fmt.Println("validatorPubKey: ", validator1PubKey)

	validator1AddressCMD := &exec.Cmd{
		Path: chainExecutable,
		Args: []string{chainExecutable, "keys", "show", "validator-1", "-a", "--keyring-backend", "test"},
	}
	fmt.Println("storing validator address in variable")
	out, err := validator1AddressCMD.CombinedOutput()
	if err != nil {
		fmt.Print("error: ", err)
	}

	// var treatedPersistentPeerString = strings.Replace(persistentPeerString, "\n", "", -1)

	validator1Address := strings.ReplaceAll(string(out), "\n", "")

	fmt.Println("getting out the address again")
	fmt.Println(validator1Address)

	addGenesisAccountValidator1Cmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "add-genesis-account", validator1Address, "100000000000stake"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	fmt.Println("adding genesis accounts to the validator")
	if err := addGenesisAccountValidator1Cmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	createGentXValidator1Cmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "gentx", "validator-1", "100000000stake", "--chain-id", "test-chain", "--keyring-backend", "test"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	fmt.Println("creating gentx for validator-1")
	if err := createGentXValidator1Cmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	e = os.Rename(dir+"/.test-chain/config/node_key.json", dir+"/.test-chain/config/node_key_1.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
	e = os.Rename(dir+"/.test-chain/config/priv_validator_key.json", dir+"/.test-chain/config/priv_validator_key_1.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}

	e = os.Rename(dir+"/.test-chain/config/node_key_2.json", dir+"/.test-chain/config/node_key.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
	e = os.Rename(dir+"/.test-chain/config/priv_validator_key_2.json", dir+"/.test-chain/config/priv_validator_key.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}

	// fmt.Println(("RUNNIG GENESIS ACCOUNT FOR VALIDATOR 2"))

	validator2AddressCMD := &exec.Cmd{
		Path: chainExecutable,
		Args: []string{chainExecutable, "keys", "show", "validator-2", "-a", "--keyring-backend", "test"},
	}
	fmt.Println("storing validator address in variable")
	out, err = validator2AddressCMD.CombinedOutput()
	if err != nil {
		fmt.Print("error: ", err)
	}

	validator2Address := strings.ReplaceAll(string(out), "\n", "")

	addGenesisAccountValidator2Cmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "add-genesis-account", validator2Address, "100000000000stake", "--keyring-backend", "test"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := addGenesisAccountValidator2Cmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	createGentXValidator2Cmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "gentx", "validator-2", "100000000stake", "--chain-id", "test-chain", "--keyring-backend", "test"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := createGentXValidator2Cmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	e = os.Rename(dir+"/.test-chain/config/node_key.json", dir+"/.test-chain/config/node_key_2.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
	e = os.Rename(dir+"/.test-chain/config/priv_validator_key.json", dir+"/.test-chain/config/priv_validator_key_2.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}

	e = os.Rename(dir+"/.test-chain/config/node_key_3.json", dir+"/.test-chain/config/node_key.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
	e = os.Rename(dir+"/.test-chain/config/priv_validator_key_3.json", dir+"/.test-chain/config/priv_validator_key.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}

	validator3AddressCMD := &exec.Cmd{
		Path: chainExecutable,
		Args: []string{chainExecutable, "keys", "show", "validator-3", "-a", "--keyring-backend", "test"},
	}
	fmt.Println("storing validator address in variable")
	out, err = validator3AddressCMD.CombinedOutput()
	if err != nil {
		fmt.Print("error: ", err)
	}

	// var treatedPersistentPeerString = strings.Replace(persistentPeerString, "\n", "", -1)

	validator3Address := strings.ReplaceAll(string(out), "\n", "")

	// fmt.Println("ADDING GENESIS ACCOUNT VALIDATOR 3 CMD")
	addGenesisAccountValidator3Cmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "add-genesis-account", validator3Address, "100000000000stake", "--keyring-backend", "test"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := addGenesisAccountValidator3Cmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	createGentXValidator3Cmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "gentx", "validator-3", "100000000stake", "--chain-id", "test-chain", "--keyring-backend", "test"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := createGentXValidator3Cmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	collectGentX()

	// fmt.Println("COLLECT TANSACTIONS")

	// collectGentXCmd := &exec.Cmd{
	// 	Path: chainExecutable,
	// 	Args: []string{chainExecutable, "collect-gentxs"},
	// }

	// if err := collectGentXCmd.Run(); err != nil {
	// 	fmt.Println("error: ", err)
	// }

}

var generateTestInfra = &cobra.Command{
	Use:   "generate-test-infra",
	Short: "Create cloud validator set",
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

		// chmodExecutable, _ := exec.LookPath("chmod")
		// chmodValidatorKeyPem := &exec.Cmd{
		// 	Path:   chmodExecutable,
		// 	Args:   []string{chmodExecutable, "400", "validator_key.pem"},
		// 	Stdout: os.Stdout,
		// 	Stderr: os.Stederr,
		// }
	},
}

// generateTestNetCmd represents the generateTestNet command
var generateTestNetCmd = &cobra.Command{
	Use:   "generate-test-net",
	Short: "One click testnet for starport scaffolded applications",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generateTestNet called")

		/*
			Assumptions -
				1. starport chain init has been run once, so .test-chain exists and a local binary exists at the gopath.
				2. Terraform exists
				3. AWS is configured
		*/

		// We first need to clear up any existing node_key.json and priv_validator_key.json
		setup()
		generateValidatorKeys(1)
		generateValidatorKeys(2)
		generateValidatorKeys(3)

		generateGenesisTransactionsAndAccounts()

		generateBuildArtifacts()
		pushToECR()

		configureValidators()
		updateValidators()

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
	// rootCmd.AddCommand((generateTestInfra))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateTestNetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateTestNetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
