#FROM golang:latest AS builder
FROM golang:latest
ADD . /backhub
WORKDIR /backhub
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o main .

RUN apt install git

# final stage
#FROM alpine:latest
#RUN apk --no-cache add ca-certificates
#COPY --from=builder /backhub ./
#RUN chmod +x ./backhub

ENTRYPOINT ["/backhub/main"]
