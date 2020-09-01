FROM alpine:3.12
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
