FROM alpine:3.13.0
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
