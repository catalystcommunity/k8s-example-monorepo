FROM golang:1.23 AS builder

WORKDIR /workspace

# create api directory, so that the docker build context contains the same
# directory structure as the repository
RUN mkdir app_api
RUN mkdir coredb

# download dependencies
COPY app_api/go.mod ./app_api
COPY app_api/go.sum ./app-api
COPY ./coredb/ ./coredb/
RUN ls -l ./
RUN ls -l coredb/
WORKDIR /workspace/app_api/
RUN go mod download

# copy source code
COPY ./app_api/ .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o main main.go

# Use distroless as minimal base image
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/app_api/main .
USER 65532:65532

ENTRYPOINT ["/main"]
