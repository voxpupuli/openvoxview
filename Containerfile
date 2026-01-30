FROM alpine:3.21 AS build

ENV GIN_MODE=release
RUN apk add --no-cache go yarn make

WORKDIR /build
COPY . /build

ARG VUE_APP_COMMIT=dirty
ENV VUE_APP_COMMIT=${VUE_APP_COMMIT}

RUN make backend

FROM alpine:3.21
ENV PORT=5000
COPY --from=build /build/openvoxview /openvoxview

ENTRYPOINT /openvoxview
