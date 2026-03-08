FROM golang:1.24.12-bookworm AS go-builder

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/worker ./cmd/worker

FROM node:24-bookworm-slim

ARG GH_VERSION=2.87.0
ARG GO_VERSION=1.24.12
ARG CODEX_VERSION=0.106.0

ENV PATH="/usr/local/go/bin:${PATH}"

RUN apt-get update && \
    apt-get install -y --no-install-recommends bash ca-certificates curl git jq make unzip && \
    curl -fsSL "https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_amd64.tar.gz" -o /tmp/gh.tgz && \
    tar -xzf /tmp/gh.tgz -C /tmp && \
    mv /tmp/gh_${GH_VERSION}_linux_amd64/bin/gh /usr/local/bin/gh && \
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf /tmp/go.tgz && \
    npm install -g "@openai/codex@${CODEX_VERSION}" && \
    rm -rf /var/lib/apt/lists/* /tmp/gh.tgz /tmp/go.tgz /tmp/gh_${GH_VERSION}_linux_amd64

WORKDIR /app
COPY --from=go-builder /out/worker /usr/local/bin/platform-ai-worker
COPY prompts ./prompts

ENTRYPOINT ["platform-ai-worker"]
CMD ["run"]
