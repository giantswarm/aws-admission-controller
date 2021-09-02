FROM alpine:3.14.2
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
