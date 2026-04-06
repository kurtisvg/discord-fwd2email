FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/fwd2email .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /bin/fwd2email /bin/fwd2email
COPY --from=build /src/LICENSE /LICENSE
ENTRYPOINT ["/bin/fwd2email"]
