# Use an official Golang runtime as a parent image
FROM --platform=linux/amd64 golang:latest as builder

# Set the working directory to /app
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# Install any needed packages specified in go.mod
RUN go mod download

# Compile the Go program for a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/ergomake cmd/ergomake/main.go

# Use an official Alpine Linux as a base image
FROM --platform=linux/amd64 alpine:latest

RUN apk update && apk add --no-cache git

# Set the working directory to /app
WORKDIR /app

# Copy the compiled binary from the builder image
COPY --from=builder /app/ergomake /app/ergomake

EXPOSE 8080
EXPOSE 9090

# Define the command to run the executable when the container starts
CMD ["/app/ergomake"]
