FROM golang:1.22.1 as builder

WORKDIR /
COPY . .
RUN go build -o bin/drone-github-commit-status --ldflags '-linkmode external -extldflags "-static"'

FROM alpine:3.17

RUN apk add --no-cache libc6-compat
COPY --from=builder /bin/drone-github-commit-status /usr/bin

CMD [ "/usr/bin/drone-github-commit-status" ]