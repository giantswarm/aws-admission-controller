FROM alpine:3.12
WORKDIR /app
COPY admission-controller /app
CMD ["/app/admission-controller"]
