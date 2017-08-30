FROM golang:1.8-onbuild AS build-env

FROM scratch
WORKDIR /app
COPY --from=build-env /go/src/app /app/
ENTRYPOINT ./app
