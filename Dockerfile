FROM golang:1.18 AS builder

WORKDIR /code/

COPY ./go.mod /code/go.mod
COPY ./go.sum /code/go.sum
RUN go mod download

COPY . /code/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make clean build


FROM debian:bookworm
RUN apt update && apt install -y ca-certificates

WORKDIR /cbr
COPY --from=builder /code/build/cbr /cbr/
COPY --from=builder /code/webroot /cbr/webroot

EXPOSE 8080/tcp

ENTRYPOINT [ "./cbr" ]
