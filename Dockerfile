FROM golang:1.23-alpine as backend

ARG GIT_BRANCH
ARG GITHUB_SHA

ENV GOFLAGS="-mod=vendor"
ENV CGO_ENABLED=0

ADD . /build
WORKDIR /build

RUN apk add --no-cache --update git tzdata ca-certificates make

RUN make linux

FROM scratch

COPY --from=backend /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend /build/build/hub.linux /hub

EXPOSE 6587
WORKDIR /srv
ENTRYPOINT ["/hub"]
