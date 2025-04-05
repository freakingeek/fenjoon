FROM golang:1.24.2

ENV PORT=8080

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd

RUN go build -o /fenjoon

EXPOSE 8080

ENTRYPOINT [ "/fenjoon" ]
