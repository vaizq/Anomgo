FROM golang:1.23 AS base

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . ./

FROM base as tool

WORKDIR /build/cmd/tool
COPY --from=base /build/uploads /uploads
CMD ["/build/cmd/tool/tool"]

FROM base as builder

WORKDIR /build/cmd
RUN go build -o ./store/store ./store
RUN go build -o ./tool/tool ./tool

FROM gcr.io/distroless/base-debian12 AS deploy

WORKDIR /
COPY --from=builder /build/cmd/store/store /
COPY --from=builder /build/cmd/tool/tool /
COPY --from=base /build/uploads /uploads
COPY --from=base /build/ui /ui
EXPOSE 4000
ENTRYPOINT ["/store"]
