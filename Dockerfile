FROM golang:alpine as builder

RUN apk update && apk add git 
WORKDIR /app
COPY ./main.go ./
RUN go mod init github.com/blackrez/openweathermap_exporter
RUN go get -d -v
RUN go build -o /openweathermap_exporter


FROM alpine
EXPOSE 2112
COPY --from=builder /openweathermap_exporter /openweathermap_exporter
ENTRYPOINT ["/openweathermap_exporter"]
