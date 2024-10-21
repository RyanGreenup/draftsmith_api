############################################################
## Build the Binary ########################################
############################################################
FROM golang:alpine as builder

RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./src ./src
RUN go build -o draftsmith_api ./src/

############################################################
## Run the Binary ##########################################
############################################################
FROM alpine:3.18
# This is handy for debugging
RUN apk add --no-cache postgresql
WORKDIR /app
COPY --from=builder /app/draftsmith_api .
CMD ./draftsmith_api --db_host=db serve
# CMD tail -f /dev/null
