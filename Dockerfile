# ref: https://github.com/jeremyhuiskamp/golang-docker-scratch
################################
# STEP 1 build executable binary
################################
ARG PROG_NAME=app
FROM golang:alpine as builder
ARG PROG_NAME
ENV PROG_NAME $PROG_NAME

ARG GOARCH
ENV GOARCH ${GOARCH}
ENV CGO_ENABLED=0

WORKDIR /src

COPY *.go .
COPY *.html .

# Static build required so that we can safely copy the binary over.
# `-tags timetzdata` embeds zone info from the "time/tzdata" package.
RUN go build -ldflags '-extldflags "-static"' -tags timetzdata -o $PROG_NAME .

################################
# STEP 2 build a small runtime image
################################
FROM alpine:3
ARG PROG_NAME
ENV PROG_NAME $PROG_NAME

# Allow changing the username at build time, default to highbyte
ARG USERNAME=converter
RUN adduser -D $USERNAME

ENV APP_HOME /app
WORKDIR $APP_HOME

RUN apk add inkscape \
        build-base \
        msttcorefonts-installer fontconfig && \
    update-ms-fonts && \
    fc-cache -f

# Don't run as root
USER $USERNAME

# Copy our static executable.
COPY --chown=$USERNAME:$USERNAME --from=builder /src/$PROG_NAME /$APP_HOME/$PROG_NAME

ENTRYPOINT [ "sh", "-c", "./$PROG_NAME -serve -port 8080" ]
