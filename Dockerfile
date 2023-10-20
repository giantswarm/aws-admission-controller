FROM alpine:3.18.4
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
