FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS TARGETARCH TAGVERSION

WORKDIR /src
COPY --from=github . .
WORKDIR /build
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -C /src/ -o /build/shortlink.app -ldflags "-X 'github.com/s5i/goutil/version.External=${TAGVERSION}'" -tags netgo,osusergo .

FROM alpine
RUN apk add bash
COPY --from=build /src/entrypoint.sh /app/entrypoint.sh
COPY --from=build /src/example_config.yaml /app/example_config.yaml
COPY --from=build /build/shortlink.app /app/shortlink.app
VOLUME /cfg
VOLUME /data
CMD [ "/app/entrypoint.sh" ]

VOLUME /appdata
ENTRYPOINT /app/entrypoint.sh
