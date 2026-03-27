FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o marketplace ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app/marketplace .
COPY ./migrations ./migrations
COPY ./openapi ./openapi
ENV PORT=8080
ENV JWT_SECRET=your_generated_secret_here
EXPOSE 8080
CMD ["./marketplace"]