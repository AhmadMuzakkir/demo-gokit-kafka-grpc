FROM golang:1.15 AS builder

WORKDIR /project

COPY . .

RUN CGO_ENABLED=0 go build -mod vendor -o server ./cmd/server/main.go

FROM alpine:3.12

RUN apk add --no-cache bash

COPY --from=builder /project/server /server
COPY --from=builder /project/wait-for-it.sh /wait-for-it.sh

CMD ["bash", "/server"]