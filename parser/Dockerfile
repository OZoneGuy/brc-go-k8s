FROM golang:alpine AS builder
WORKDIR /src/app
COPY ./ ./
RUN go build -o main

FROM gcr.io/distroless/static AS final

COPY --from=builder --chown=nonroot:nonroot /src/app/main /main

ENTRYPOINT ["/main"]
