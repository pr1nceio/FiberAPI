FROM golang:1.22 AS builder
RUN mkdir /app
WORKDIR /app
ADD go.mod go.sum /app/
RUN echo "Resolving deps..." && \
    go install github.com/swaggo/swag/cmd/swag@latest && \
    go install github.com/gordonklaus/ineffassign@latest && \
    go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
ADD . /app
RUN bash build.sh

FROM alpine AS runner
RUN mkdir /app
RUN apk add --no-cache tzdata
COPY --from=builder /app/FiberAPI /app
CMD ["/app/FiberAPI"]