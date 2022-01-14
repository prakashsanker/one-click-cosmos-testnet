FROM golang:latest

WORKDIR /

COPY ./dist /dist

EXPOSE 26657
EXPOSE 1317


# CMD ["bash","/dist/start.sh"]

CMD ["/dist/test-chaind", "start", "--home", "/dist/.test-chain" ]







