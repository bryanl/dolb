SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

AGENTBIN=dolb-agent

$(AGENTBIN): $(SOURCES)
	GOOS=linux GOATCH=amd64 go build -o cmd/dolb-agent/${AGENTBIN} github.com/bryanl/dolb/cmd/dolb-agent

.PHONY: deploy-agent
deploy-agent: $(AGENTBIN)
	docker build -f Dockerfile.agent -t bryanl/dolb-agent:0.0.2 . && docker push bryanl/dolb-agent:0.0.2
