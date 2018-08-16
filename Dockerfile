FROM golang:1

ENV CONFIG_PATH="/etc/japanbot-go"

WORKDIR $GOPATH/src/github.com/hakasec/japanbot-go
COPY . .

# Install dep and dependencies
RUN [ "go", "get", "-u", "github.com/golang/dep/cmd/dep" ]
RUN [ "dep", "ensure" ]

# Build and install
RUN [ "go", "install" ]

# Mount points
VOLUME [ ${CONFIG_PATH} ]

WORKDIR ${CONFIG_PATH}
CMD [ "japanbot-go" ]