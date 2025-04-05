FROM golang:1.24.2

ENV PORT 8080

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /fenjoon

EXPOSE 8080

ENTRYPOINT [ "/fenjoon" ]