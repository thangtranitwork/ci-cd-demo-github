# Bước 1 (Build Stage): Dùng môi trường Go chuẩn để biên dịch mã nguồn
FROM golang:1.21-alpine AS builder

# Cài git (cần thiết cho go mod download một số package)
RUN apk add --no-cache git

# Thiết lập thư mục làm việc bên trong Container
WORKDIR /app

# Copy file cấu hình module trước để tận dụng bộ nhớ đệm (Cache) của Docker
COPY go.mod go.sum ./
RUN go mod download

# Tải toàn bộ mã nguồn vào
COPY . .

# Biên dịch mã nguồn thành file chạy (Tắt CGO để file chạy độc lập)
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Bước 2 (Run Stage): Chuyển sang một image cực nhẹ để vận hành
FROM alpine:latest

# Cài ca-certificates để hỗ trợ HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy file 'main' đã biên dịch thành công từ bước 1 sang đây
COPY --from=builder /app/main .

# Khai báo với Docker là app này sẽ lắng nghe ở cổng 8080
EXPOSE 8080

# Chạy thực thi
CMD ["./main"]
