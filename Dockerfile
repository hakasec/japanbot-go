FROM golang:1

ENV CONFIG_PATH="/etc/japanbot-go"

WORKDIR $GOPATH/src/github.com/hakasec/japanbot-go
COPY . .

# Install dependencies
RUN [                                   \
    "go", "get", "-u",                  \
    "github.com/bwmarrin/discordgo",    \
    "github.com/hakasec/jmdict-go"      \
]

# Build and install
RUN [ "go", "install" ]

# Mount points
VOLUME [ ${CONFIG_PATH} ]

WORKDIR ${CONFIG_PATH}
CMD [ "japanbot-go" ]