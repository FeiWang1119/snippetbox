FROM golang AS development

WORKDIR /snippetbox

COPY /cmd/ /snippetbox/cmd
COPY /pkg/ /snippetbox/pkg
COPY /ui/ /snippetbox/ui
COPY /tls/ /snippetbox/tls

RUN go mod init snippetbox
RUN go mod tidy
RUN go build -o /snippetbox/snippetbox ./cmd/web/*.go

FROM debian AS builder

WORKDIR /

COPY --from=development /snippetbox/snippetbox /usr/bin/snippetbox
COPY /ui /ui
COPY /tls /tls

EXPOSE 8888

CMD ["snippetbox", "-addr=:8888"]
