FROM alpine
EXPOSE 8080
RUN mkdir /app
COPY FiberAPI /app
CMD ["/app/FiberAPI"]