package testnet

import (
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
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

var nodeIdsArray []string

func rmGenesisFile() {
	rmExecutable, _ := exec.LookPath("rm")
	usr, _ := user.Current()
	dir := usr.HomeDir
	rmGenesisCmd := &exec.Cmd{
		Path:   rmExecutable,
		Args:   []string{rmExecutable, dir + "/.test-chain/config/genesis.json"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := rmGenesisCmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}
}

func GenerateValidatorKeys(validatorNumber int64) {

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

	if err != nil {
		fmt.Println("error: ", err)
	}

	rmGenesisFile()

	initCmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "init", "validator-" + validatorNumberStr, "--chain-id", "test-chain"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := initCmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}
	nodeIdCmd := &exec.Cmd{
		Path: chainExecutable,
		Args: []string{chainExecutable, "tendermint", "show-node-id"},
	}

	out, err = nodeIdCmd.CombinedOutput()

	nodeIdsArray = append(nodeIdsArray, string(out))

	e := os.Rename(dir+"/.test-chain/config/node_key.json", dir+"/.test-chain/config/node_key_"+validatorNumberStr+".json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
	e = os.Rename(dir+"/.test-chain/config/priv_validator_key.json", dir+"/.test-chain/config/priv_validator_key_"+validatorNumberStr+".json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
}

func GenerateBuildArtifacts() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	dockerExecutable, _ := exec.LookPath("docker")
	os.Chdir(dir + "/test-chain")

	rmDistCMD := exec.Command("rm", "-rf", dir+"/test-chain/dist")

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

	// we need to temporarily move the start script over
	cpExecutable, _ := exec.LookPath("cp")

	moveStartScript := &exec.Cmd{
		Path:   cpExecutable,
		Args:   []string{cpExecutable, dir + "/one-click-cosmos-testnet/start.sh", dir + "/test-chain/dist/"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := moveStartScript.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	latestSha := getLatestSha()

	buildDockerImage := &exec.Cmd{
		Path:   dockerExecutable,
		Args:   []string{dockerExecutable, "buildx", "build", "--platform", "linux/amd64", "-t", latestSha, "-f", dir + "/one-click-cosmos-testnet/Dockerfile", dir + "/test-chain", "-t", "test-chain", "--no-cache"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := buildDockerImage.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	tagDockerImage := &exec.Cmd{
		Path:   dockerExecutable,
		Args:   []string{dockerExecutable, "tag", "test-chain:latest", "187926495729.dkr.ecr.ap-south-1.amazonaws.com/one-click-cosmos-testnet-repo:" + latestSha},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := tagDockerImage.Run(); err != nil {
		fmt.Println("error: ", err)
	}
}

func UpdateValidators() {
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

	err := os.Mkdir(dir+"/.test-chain/config/validator-config", 0770)
	if err != nil {
		fmt.Println(err)
	}

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

	copyConfig := &exec.Cmd{
		Path:   scpExecutable,
		Args:   []string{scpExecutable, "-o StrictHostKeyChecking=no", "-o IdentitiesOnly=yes", "-i", dir + "/one-click-cosmos-testnet/validator_key.pem", "-r", dir + "/.test-chain/config/validator-config", "ec2-user@" + dnsName + ":~/"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := copyConfig.Run(); err != nil {
		fmt.Println("Scp error: ", err)
	}

	copyConfigTomlCmd := &exec.Cmd{
		Path:   scpExecutable,
		Args:   []string{scpExecutable, "-o StrictHostKeyChecking=no", "-o IdentitiesOnly=yes", "-i", dir + "/one-click-cosmos-testnet/validator_key.pem", "-pr", dir + "/.test-chain/config/config.toml", "ec2-user@" + dnsName + ":~/validator-config"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := copyConfigTomlCmd.Run(); err != nil {
		fmt.Print("Config.toml error: ", err)
	}

	copyStartScriptCMD := &exec.Cmd{
		Path:   scpExecutable,
		Args:   []string{scpExecutable, "-o StrictHostKeyChecking=no", "-o IdentitiesOnly=yes", "-i", dir + "/one-click-cosmos-testnet/validator_key.pem", "-pr", dir + "/one-click-cosmos-testnet/start.sh", "ec2-user@" + dnsName + ":~/validator-config"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := copyStartScriptCMD.Run(); err != nil {
		fmt.Print("CP Start Script error: ", err)
	}

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

func ConfigureValidators() {
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
		usr, _ := user.Current()
		dir := usr.HomeDir

		persistentPeerString := constructPersistentPeerString(instances)

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

		setAddrBookStrictToFalseCMD := &exec.Cmd{
			Path:   sedExecutable,
			Args:   []string{sedExecutable, "-i", "''", "s/addr_book_strict = true/addr_book_strict = false/g", dir + "/.test-chain/config/config.toml"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := setAddrBookStrictToFalseCMD.Run(); err != nil {
			fmt.Println("error: ", err)
		}

		enableAPIServerCMD := &exec.Cmd{
			Path:   sedExecutable,
			Args:   []string{sedExecutable, "-i", "''", "s/enable = false/enable = true/g", dir + "/.test-chain/config/config.toml"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := enableAPIServerCMD.Run(); err != nil {
			fmt.Println("error: ", err)
		}

		for i, instance := range instances {
			dnsName := instance.DnsName
			moveConfigIntoValidatorConfigFolder(dnsName, i+1)
		}

	}
}

func Info(str string) {
	fmt.Println("Info: ", str)
}

func CheckIfError(err error) {
	if err != nil {
		fmt.Println("error: ", err)
	}
}

func getLatestSha() string {
	/*
		1. Build image + tag with github SHA.
		2. Push image
		3. Update using ECS
	*/

	usr, _ := user.Current()
	dir := usr.HomeDir
	// Clones the given repository, creating the remote, the local branches
	// and fetching the objects, everything in memory:
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: dir + "/test-chain/",
	})
	CheckIfError(err)

	// Gets the HEAD history from HEAD, just like this command:
	Info("git log")

	// ... retrieves the branch pointed by HEAD
	ref, err := r.Head()
	fmt.Println(ref)
	CheckIfError(err)

	return ref.Hash().String()
	// now we want to use this SHA and build the docker image
}

func PushToEcr() {
	awsExecutable, _ := exec.LookPath("aws")
	dockerExecutable, _ := exec.LookPath("docker")

	ecrGetCredentialsCMD := &exec.Cmd{
		Path: awsExecutable,
		Args: []string{awsExecutable, "ecr", "get-login-password", "--region", "ap-south-1"},
	}

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

	latestSha := getLatestSha()

	dockerPushECRCMD := &exec.Cmd{
		Path:   dockerExecutable,
		Args:   []string{dockerExecutable, "push", "187926495729.dkr.ecr.ap-south-1.amazonaws.com/one-click-cosmos-testnet-repo:" + latestSha},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := dockerPushECRCMD.Run(); err != nil {
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

	dockerPushLatestECRCMD := &exec.Cmd{
		Path:   dockerExecutable,
		Args:   []string{dockerExecutable, "push", "187926495729.dkr.ecr.ap-south-1.amazonaws.com/one-click-cosmos-testnet-repo:latest"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := dockerPushLatestECRCMD.Run(); err != nil {
		fmt.Println("error: ", err)
	}
}

func Setup() {
	// clean up
	usr, _ := user.Current()
	dir := usr.HomeDir

	rmNodeKeyCMD := exec.Command("rm", "-f", dir+"/.test-chain/config/node_key.json")

	if err := rmNodeKeyCMD.Run(); err != nil {
		fmt.Println("error: ", err)

	}

	rmValidatorKeyCMD := exec.Command("rm", "-f", dir+"/.test-chain/config/priv_validator_key.json")

	if err := rmValidatorKeyCMD.Run(); err != nil {
		fmt.Println("error: ", err)

	}

	rmGenesisJsonCMD := exec.Command("rm", "-f", dir+"/.test-chain/config/genesis.json")
	if err := rmGenesisJsonCMD.Run(); err != nil {
		fmt.Println("error: ", err)
	}
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

func GenerateGenesisTransactionsAndAccounts() {
	chainExecutable, _ := exec.LookPath("test-chaind")
	fmt.Println("generating genesis transactions and accounts")

	usr, _ := user.Current()
	dir := usr.HomeDir
	e := os.Rename(dir+"/.test-chain/config/node_key_1.json", dir+"/.test-chain/config/node_key.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
	e = os.Rename(dir+"/.test-chain/config/priv_validator_key_1.json", dir+"/.test-chain/config/priv_validator_key.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}

	validator1AddressCMD := &exec.Cmd{
		Path: chainExecutable,
		Args: []string{chainExecutable, "keys", "show", "validator-1", "-a", "--keyring-backend", "test"},
	}
	out, err := validator1AddressCMD.CombinedOutput()
	if err != nil {
		fmt.Print("error: ", err)
	}

	// var treatedPersistentPeerString = strings.Replace(persistentPeerString, "\n", "", -1)

	validator1Address := strings.ReplaceAll(string(out), "\n", "")

	addGenesisAccountValidator1Cmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "add-genesis-account", validator1Address, "100000000000stake"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := addGenesisAccountValidator1Cmd.Run(); err != nil {
		fmt.Println("error: ", err)
	}

	createGentXValidator1Cmd := &exec.Cmd{
		Path:   chainExecutable,
		Args:   []string{chainExecutable, "gentx", "validator-1", "100000000stake", "--chain-id", "test-chain", "--keyring-backend", "test"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
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

	validator2AddressCMD := &exec.Cmd{
		Path: chainExecutable,
		Args: []string{chainExecutable, "keys", "show", "validator-2", "-a", "--keyring-backend", "test"},
	}
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

	validator3Address := strings.ReplaceAll(string(out), "\n", "")

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

	e = os.Rename(dir+"/.test-chain/config/node_key.json", dir+"/.test-chain/config/node_key_3.json")
	if e != nil {
		fmt.Println("rename error: ", e)
	}
	e = os.Rename(dir+"/.test-chain/config/priv_validator_key.json", dir+"/.test-chain/config/priv_validator_key_3.json")
	if e != nil {
		fmt.Println("rename error: ", e)
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
