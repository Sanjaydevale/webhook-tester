FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o main ./cmd/server/main.go


FROM alpine
WORKDIR /app
COPY --from=build /app/main .

EXPOSE 8080
CMD [ "/app/main" ]