FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/rundown-engine ./cmd/worker

# ---

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/rundown-engine /usr/local/bin/rundown-engine
COPY DOCUMENTATION.md /app/DOCUMENTATION.md

WORKDIR /app

ENV RUNDOWN_HOST=0.0.0.0
ENV RUNDOWN_PORT=8181
ENV RUNDOWN_STORE=sqlite

EXPOSE 8181

ENTRYPOINT ["rundown-engine"]
