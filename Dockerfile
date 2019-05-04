FROM golang:1.12.2-stretch AS BASE
RUN apt-get update && apt-get install -y \
  curl \
  jq \
  ca-certificates

# Download go-swagger tool
RUN download_url=$(curl -s "https://api.github.com/repos/go-swagger/go-swagger/releases/latest" |\
  jq -r '.assets[] | select(.name | contains("'"$(uname | tr '[:upper:]' '[:lower:]')"'_amd64")) | .browser_download_url') &&\
  curl -o /usr/local/bin/swagger -L'#' "$download_url" &&\
  chmod +x /usr/local/bin/swagger

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN make -B gen
RUN make build

FROM debian:stretch
COPY --from=BASE /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=BASE /go/src/app/bin /usr/local/bin
ENTRYPOINT ["/usr/local/bin/feedgen"]