FROM alpine:3.13.4
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
