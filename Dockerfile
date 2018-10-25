FROM golang:1.11-alpine
ARG target
ADD . /go/src/github.com/zoncoen-sample/k8s-grpc-client-side-lb
RUN go install github.com/zoncoen-sample/k8s-grpc-client-side-lb/cmd/$target
RUN mv /go/bin/$target /go/bin/app

FROM alpine:latest
COPY --from=0 /go/bin/app .
CMD ./app
