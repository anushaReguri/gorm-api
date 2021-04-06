FROM golang
COPY . /gorm-api
ENTRYPOINT gorm-api
EXPOSE 8080
CMD go run main.go


