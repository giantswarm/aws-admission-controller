FROM alpine:3.13.5
WORKDIR /app
COPY aws-admission-controller /app
CMD ["/app/aws-admission-controller"]
