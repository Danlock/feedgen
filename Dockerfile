FROM golang:1.12.2-stretch AS BASE
RUN apt-get update && apt-get install -y \
  curl \
  jq \
  ca-certificates

# Download go-swagger tool
RUN download_url=$(curl -s "https://api.github.com/repos/go-swagger/go-swagger/releases/latest" | \
  jq -r '.assets[] | select(.name | contains("'"$(uname | tr '[:upper:]' '[:lower:]')"'_amd64")) | .browser_download_url') &&\
  curl -o /usr/local/bin/swagger -L'#' "$download_url" &&\
  chmod +x /usr/local/bin/swagger

WORKDIR /go/src/github.com/danlock/feedgen
COPY ./design ./design
COPY ./gen ./gen
RUN make -B gen
COPY . .
RUN make version
RUN make build

FROM scratch
COPY --from=BASE /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=BASE /go/src/github.com/danlock/feedgen/bin /usr/local/bin
COPY --from=BASE /go/src/github.com/danlock/feedgen/ui /usr/local/etc/feedgen/ui
ENTRYPOINT ["/usr/local/bin/feedgen"]