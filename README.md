## Introduction

This is an attempt to make it easier to get started with Cosmos SDK. This CLI will -

1. Create 3 EC2 instances that are validators that come to consensus and commit transactions.
2. Allow you to deploy your cosmos sdk code easily.
3. Allow you to submit transactions against a real life blockchain, in order to simulate a production environment.
4. Gives you a clean testing environment against which to run your code.

Please note that this repo is WIP. Please file an issue if you come across an error.

## Prerequisites

You will need the following tech to run this cli

1. [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli)
2. [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
3. [Docker](https://docs.docker.com/get-docker/).
4. [GoLang](https://go.dev/doc/install)
5. [Starport](https://docs.starport.com/guide/install.html)

Make _sure_ that you have Docker desktop running. Otherwise this will not work.

## Using this CLI

1. Clone this repo
2. Run `go install github.com/prakashsanker/one-click-cosmos-testnet`
3. Scaffold a chain with starport `starport scaffold chain <chain-name>`
4. Run `starport chain serve`. Once you see that the chain is running, quit the process.
5. Run `one-click-cosmos-testnet configure` and enter in your AWS details. To generate an access key and a secret key follow [this](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html) documentation.
6. Run `one-click-cosmos-testnet generate-test-infra`
7. Run `one-click-cosmos-testnet generate-test-net`

You will see output like so

```
Node - 0
SSH into Node0 with this command ssh -i 'validator_key.pem' -o IdentitiesOnly=yes ec2-user@ec2-65-0-204-22.ap-south-1.compute.amazonaws.com
Submit transactions at ec2-65-0-204-22.ap-south-1.compute.amazonaws.com:1317
Node - 1
SSH into Node1 with this command ssh -i 'validator_key.pem' -o IdentitiesOnly=yes ec2-user@ec2-3-110-92-148.ap-south-1.compute.amazonaws.com
Submit transactions at ec2-3-110-92-148.ap-south-1.compute.amazonaws.com:1317
Node - 2
SSH into Node2 with this command ssh -i 'validator_key.pem' -o IdentitiesOnly=yes ec2-user@ec2-65-0-27-30.ap-south-1.compute.amazonaws.com
Submit transactions at ec2-65-0-27-30.ap-south-1.compute.amazonaws.com:1317
```

The url at port `1317` is where you can submit transactions.

### Deploys

Right now, this is hacky, but works.

1. Run `one-click-cosmos-testnet deploy`

Rollbacks are not supported yet, but you can approximate by checking out (in git) and then running the deploy command.

### Getting Validator Information

Run `one-click-cosmos-testnet get-validators`.

### Diagnosing problems

You can ssh into the EC2 instance with the provided ssh command from `get-validators`.

Once there, run `docker container ls` and then `docker exec -it <CONTAINER_ID> /bin/bash` to ssh into the docker container.
