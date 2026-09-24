# syntax=docker/dockerfile:1
FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/provider ./cmd/provider

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/provider /provider
USER 65532:65532
ENTRYPOINT ["/provider"]
