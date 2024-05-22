FROM golang:1.22.2 as builder
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN bash build.sh

FROM alpine as runner
RUN mkdir /app
RUN apk add --no-cache tzdata
COPY --from=builder /app/FiberAPI /app
CMD ["/app/FiberAPI"]