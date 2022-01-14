#!/bin/bash

sleep 30s 
# TODO: Figure out why the CMD is running before the bind mount is ready

cp /validator-config/node_key.json /dist/.test-chain/config/node_key.json &&
cp /validator-config/priv_validator_key.json /dist/.test-chain/config/priv_validator_key.json &&
cp /validator-config/config.toml /dist/.test-chain/config/config.toml &&
/dist/test-chaind start --home /dist/.test-chain

