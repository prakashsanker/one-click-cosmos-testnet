FROM starport/cli:latest

COPY . /app

WORKDIR /app

RUN ["starport", "chain", "build"]

RUN ["/go/bin/test-chaind", "init", "validator-1"]
