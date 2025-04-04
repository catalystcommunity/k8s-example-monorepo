FROM golang:1.23 as builder

WORKDIR /workspace

# create api directory, so that the docker build context contains the same
# directory structure as the repository
RUN mkdir app_api
WORKDIR /workspace/app_api/

# copy protos from protos context, requires docker build to be ran with
# --build-context protos=../protos-api
# this is so that the go.mod file can continue to locally reference the protos
# in the docker image
COPY --from=protos . ./protos

# download dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o main main.go

# Use distroless as minimal base image
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/api/main .
USER 65532:65532

ENTRYPOINT ["/main"]
