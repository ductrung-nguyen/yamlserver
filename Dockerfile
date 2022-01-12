ARG GOLANG_VER=1.17.5
FROM dockerhub.rnd.amadeus.net:5002/golang:${GOLANG_VER}-alpine3.15 as builder

WORKDIR /go/src/app

RUN apk add --update gcc libc-dev

############## Download required golang packages
COPY ./go.* ./
RUN go mod download

COPY ./main.go .

RUN CGO_ENABLED=0 go build -tags=jsoniter

FROM dockerhub.rnd.amadeus.net:5002/alpine:3.15 as production
COPY --from=builder /go/src/app/yamlserver .

CMD ["./yamlserver"]