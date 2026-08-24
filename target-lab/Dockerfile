FROM golang:1.22-alpine AS build
WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/lab .

FROM alpine:3.20
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
 && apk add --no-cache tzdata wget
ENV TZ=Asia/Shanghai
COPY --from=build /out/lab /app/lab
EXPOSE 8090
HEALTHCHECK --interval=5s --timeout=3s --retries=8 CMD wget -qO- http://127.0.0.1:8090/health || exit 1
CMD ["/app/lab"]
