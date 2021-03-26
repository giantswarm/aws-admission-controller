FROM alpine:3.13.3
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
