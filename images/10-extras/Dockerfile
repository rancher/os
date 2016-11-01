FROM alpine:3.4
# FROM arm64=skip arm=skip
RUN apk --purge --no-cache add kmod bash openssl
COPY extra.sh  /usr/bin
CMD ["extra.sh"]
