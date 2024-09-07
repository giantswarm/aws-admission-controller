FROM alpine:3.20.3
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
