FROM alpine:3.13.1
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
