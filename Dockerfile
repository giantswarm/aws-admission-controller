FROM alpine:3.12
WORKDIR /app
COPY g8s-admission-controller /app
CMD ["/app/g8s-admission-controller"]
