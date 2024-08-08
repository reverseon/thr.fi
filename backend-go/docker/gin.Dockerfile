# Use the official Golang image as the base image
FROM golang:1.22.6-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# env
COPY .env .env

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o main .

# Expose port 6173 to the outside world
EXPOSE 6173

# Command to run the executable
CMD ["./main"]