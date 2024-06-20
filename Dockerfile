FROM alpine:3.20.1
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
