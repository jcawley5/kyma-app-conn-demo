FROM golang:1.13 as builder

ENV GO111MODULE=on

WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY assets               ./assets
COPY internal               ./internal
COPY pkg               ./pkg
COPY cmd               ./cmd

RUN ls /app/
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o kyma-app-conn-demo ./cmd/kyma-app-conn-demo

FROM scratch
WORKDIR /app
COPY --from=builder /app/kyma-app-conn-demo /app/

EXPOSE 8080
ENTRYPOINT ["/app/kyma-app-conn-demo"]