FROM golang:latest

WORKDIR /

COPY ./dist /dist
COPY .test-chain /dist/.test-chain

VOLUME ["/validator-config"]

EXPOSE 26657
EXPOSE 1317

CMD ["/dist/test-chaind", "start", "--home", "/dist/.test-chain" ]




