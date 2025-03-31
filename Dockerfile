# using official golang's image as builder image
FROM golang:1.24 AS builder

#setting the working dir
WORKDIR /app

#copy go mod and download dependencies
COPY go.mod go.sum ./
RUN go mod download

#copy the rest of the app source code
COPY . .

#Building the GO app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd

#using a light weight image to run the app
FROM alpine:latest

#Setting the working dir in the new image
WORKDIR /root/

# Copy the .env file into the container
COPY .env .env


#copy the compiled binaries from the builder
COPY --from=builder /app/main .

# Ensure the binary is executable
RUN chmod +x ./main

#exposing the app to 8000 port
EXPOSE 8000

#run the binary when the container starts
CMD ["./main"]
