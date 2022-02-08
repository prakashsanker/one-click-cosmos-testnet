package config

import (
	"fmt"
	"os"
)

var Dockerfile = `
FROM golang:latest

WORKDIR /

COPY ./dist /dist

EXPOSE 26657
EXPOSE 1317


CMD ["bash","/dist/start.sh"]

# CMD ["/dist/test-chaind", "start", "--home", "/dist/.test-chain" ]
`

var StartScript = `
#!/bin/bash

sleep 30s 
# TODO: Figure out why the CMD is running before the bind mount is ready

cp /validator-config/node_key.json /dist/.test-chain/config/node_key.json &&
cp /validator-config/priv_validator_key.json /dist/.test-chain/config/priv_validator_key.json &&
cp /validator-config/config.toml /dist/.test-chain/config/config.toml &&
/dist/test-chaind start --home /dist/.test-chain

`

func GenerateDockerFile() {
	dockerFileTemplate, err := os.Create("./Dockerfile")

	if err != nil {
		fmt.Println(err)
	}

	defer dockerFileTemplate.Close()

	_, err = dockerFileTemplate.WriteString(Dockerfile)

	if err != nil {
		fmt.Println(err)
	}

	dockerFileTemplate.Sync()
}

func GenerateStartScript() {
	startScriptTemplate, err := os.Create("./dist/start.sh")

	if err != nil {
		fmt.Println(err)
	}

	defer startScriptTemplate.Close()

	_, err = startScriptTemplate.WriteString(StartScript)

	if err != nil {
		fmt.Println(err)
	}

	startScriptTemplate.Sync()

}
