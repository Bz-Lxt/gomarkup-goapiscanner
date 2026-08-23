FROM golang:1.22-alpine AS build
WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
 && apk add --no-cache git ca-certificates
COPY go.mod go.sum* ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
 && apk add --no-cache ca-certificates tzdata wget \
 && wget -qO /usr/share/fonts/SimHei.ttf \
    https://github.com/StellarCN/scp_zh/raw/master/fonts/SimHei.ttf || true
ENV TZ=Asia/Shanghai
ENV PDF_FONT=/usr/share/fonts/SimHei.ttf
WORKDIR /app
COPY --from=build /out/server /app/server
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --retries=8 CMD wget -qO- http://127.0.0.1:8080/api/v1/health || exit 1
CMD ["/app/server"]
