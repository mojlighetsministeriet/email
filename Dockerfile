# Run the build
FROM mojlighetsministeriet/go-polymer-faster-build
ENV WORKDIR /go/src/github.com/mojlighetsministeriet/email
COPY . $WORKDIR
WORKDIR $WORKDIR
RUN go get -t -v ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o email-amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o email-arm64

# Create the final docker image
FROM scratch
COPY --from=0 /go/src/github.com/mojlighetsministeriet/email/email-amd64 /
COPY --from=0 /go/src/github.com/mojlighetsministeriet/email/email-arm64 /
COPY run.sh /
ENTRYPOINT ["/run.sh"]
