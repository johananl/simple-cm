FROM golang:alpine as builder

RUN apk update && apk add curl git
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY . src/github.com/johananl/simple-cm
WORKDIR src/github.com/johananl/simple-cm/cmd/worker

RUN dep ensure
RUN go build -o /tmp/worker

FROM alpine

COPY --from=builder /tmp/worker /worker
COPY ./modules /etc/simple-cm/modules
EXPOSE 8888
CMD /worker