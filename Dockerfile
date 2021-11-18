FROM golang:1.14-alpine
WORKDIR /src
COPY . . 
RUN go build -o /app .

FROM vimagick/youtube-dl
COPY --from=0 /app /app
ENTRYPOINT ["/app"]