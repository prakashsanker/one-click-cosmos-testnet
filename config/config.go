package config

import (
	"fmt"
	"os"
	"strings"
)

func getChainConfigFolderName() string {
	return "." + getChainFolderName()
}

func getChainFolderName() string {
	path, _ := os.Getwd()

	splitPath := strings.Split(path, "/")
	fileName := splitPath[len(splitPath)-1]

	return fileName
}

func getChainBinaryName() string {
	return getChainFolderName() + "d"
}

var Dockerfile = `
FROM golang:latest

WORKDIR /

COPY ./dist /dist

EXPOSE 26657
EXPOSE 1317


CMD ["bash","/dist/start.sh"]
`

var StartScript = `
#!/bin/bash

sleep 30s 
# TODO: Figure out why the CMD is running before the bind mount is ready

cp /validator-config/node_key.json /dist/%s/config/node_key.json &&
cp /validator-config/priv_validator_key.json /dist/%s/config/priv_validator_key.json &&
cp /validator-config/config.toml /dist/%s/config/config.toml &&
cp /validator-config/app.toml /dist/%s/config/app.toml &&
/dist/%s start --home /dist/%s

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
	chainConfigFolderName := getChainConfigFolderName()
	chainBinaryName := getChainBinaryName()
	defer startScriptTemplate.Close()

	treatedScriptTemplate := fmt.Sprintf(StartScript, chainConfigFolderName, chainConfigFolderName, chainConfigFolderName, chainConfigFolderName, chainBinaryName, chainConfigFolderName)

	_, err = startScriptTemplate.WriteString(treatedScriptTemplate)

	if err != nil {
		fmt.Println(err)
	}

	startScriptTemplate.Sync()
}
