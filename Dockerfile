FROM golang:1.18 as builder

WORKDIR /app
COPY go.*  ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o cmd/main ./cmd


######## Start a new stage from scratch #######
FROM golang:1.18
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/cmd/main .

# Command to run the executable
CMD [ "/root/main" ]