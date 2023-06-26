FROM alpine
EXPOSE 8080
RUN apk add --no-cache tzdata
RUN mkdir /app
COPY FiberAPI /app
CMD ["/app/FiberAPI"]