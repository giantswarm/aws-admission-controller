FROM alpine:3.15.0
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
