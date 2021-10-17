FROM golang:alpine as build

# Inject env vars
ARG PAT

# Download and use git
RUN apk add git
RUN git config --global url.https://${_PAT}@github.com/.insteadOf https://github.com/

# Set necessary env variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPRIVATE=github.com/ThinkSystemIO

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o main .

# Create final container to hide history
FROM golang:alpine

# Env vars
ARG ACCOUNT
ARG KEY

# Move to working directory /dist
WORKDIR /dist

# Decode key to json file
RUN echo ${KEY} | base64 -d > /dist/auth.json


# Update packages
RUN apk update

# Add packages
RUN apk add bash
RUN apk add curl
RUN apk add openssl
RUN apk add python3

# Install helm
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
RUN chmod 700 get_helm.sh
RUN bash ./get_helm.sh

# Install gcloud client
RUN curl https://sdk.cloud.google.com > install.sh
RUN bash install.sh --disable-prompts --install-dir=/dist
RUN bash /dist/google-cloud-sdk/bin/gcloud auth activate-service-account ${ACCOUNT} --key-file=/dist/auth.json

# Copy binary from build to main folder
COPY --from=build /build/main .

# Export necessary port
EXPOSE 80

# Command to run when starting the container
CMD ["/dist/main"]