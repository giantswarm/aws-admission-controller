FROM alpine:3.13.2
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
