FROM alpine:3.14.0
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
