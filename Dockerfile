FROM golang:1.24-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o myapp .
RUN ls -l /app  # ✅ 查看构建后是否存在 binary

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/myapp .

RUN ls -l /app  # ✅ 再次确认 copy 成功

ENV PORT=8080
EXPOSE 8080
CMD ["./myapp"]
