FROM golang:1.12

# TODO: Release tracking?
# ARG VERSION="dirty"
# -ldflags "-X github.com/tokend/erc20-deposit-svc/config.Release=${VERSION}" 

WORKDIR /go/src/github.com/tokend/erc20-deposit-svc
COPY . .
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build -o /usr/local/bin/erc20-deposit-svc github.com/tokend/erc20-deposit-svc

###

FROM alpine:3.9

COPY --from=0 /usr/local/bin/erc20-deposit-svc /usr/local/bin/erc20-deposit-svc
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["erc20-deposit-svc", "run", "deposit"]

