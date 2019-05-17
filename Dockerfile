FROM golang:1

ENV SHARE_PATH="/usr/share/japanbot-go"
ENV CONFIG_PATH="/etc/japanbot-go"

WORKDIR $GOPATH/src/github.com/hakasec/japanbot-go
COPY . .

RUN mkdir ${SHARE_PATH}
RUN mkdir ${CONFIG_PATH}

# copy colours file to share
RUN cp colours.json ${SHARE_PATH}/

# Install dep and dependencies
RUN [ "go", "get", "-u", "github.com/golang/dep/cmd/dep" ]
RUN [ "dep", "ensure" ]

# Build and install
RUN [ "go", "install" ]

# Mount points
VOLUME [ ${CONFIG_PATH} ]

CMD [ "japanbot-go" ]