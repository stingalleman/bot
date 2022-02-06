FROM golang:alpine

# ENV DISCORD_TOKEN=""
ENV DB_PATH="/db.sqlite"

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /bot

CMD [ "/bot" ]
