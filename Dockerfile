FROM alpine:3.17.3
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
