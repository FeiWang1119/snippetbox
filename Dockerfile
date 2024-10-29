FROM golang AS development

WORKDIR /snippetbox

COPY /src/ /snippetbox/src

RUN go mod init snippetbox
RUN go mod tidy
RUN go build -o /snippetbox/snippetbox ./src/web/*.go


FROM debian AS builder

WORKDIR /

COPY --from=development /snippetbox/snippetbox /usr/bin/snippetbox
COPY /src/ui /ui

EXPOSE 8888

CMD ["snippetbox", "-addr=:8888"]
