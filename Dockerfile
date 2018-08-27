FROM golang:1.10 AS builder

ADD https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 /usr/bin/dep
RUN echo "287b08291e14f1fae8ba44374b26a2b12eb941af3497ed0ca649253e21ba2f83" /usr/bin/dep | sha256sum -c \
  && chmod 755 /usr/bin/dep

WORKDIR $GOPATH/src/github.com/mhemeryck/modbridge
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /app .

FROM scratch
COPY --from=builder /app ./
ENTRYPOINT ["./app"]
