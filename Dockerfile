FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS backend-builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/post-sync ./cmd/server

FROM alpine:3.21 AS backend-runtime

WORKDIR /app
RUN adduser -D -u 10001 appuser
COPY --from=backend-builder /out/post-sync /app/post-sync
USER appuser
EXPOSE 8080
CMD ["/app/post-sync"]
