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
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o kyma-app-connector./cmd/kyma-app-connector

FROM scratch
WORKDIR /app
COPY --from=builder /app/kyma-app-connector /app/

EXPOSE 8080
ENTRYPOINT ["/app/kyma-app-connector"]