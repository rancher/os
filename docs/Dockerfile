FROM rancher/rancher.github.io:build AS builder

FROM nginx
COPY --from=builder /build/_site /usr/share/nginx/html/docs
COPY --from=builder /build/favicon.png /usr/share/nginx/html/favicon.png
