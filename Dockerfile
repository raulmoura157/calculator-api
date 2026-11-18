FROM golang:1.26.10-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /calculator-api .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /calculator-api /calculator-api
EXPOSE 8000
USER 65532:65532
ENTRYPOINT ["/calculator-api"]
