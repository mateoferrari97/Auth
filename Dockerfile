FROM golang:1.15

ARG PRIVATE_KEY
ENV PRIVATE_KEY=$PRIVATE_KEY

ARG GOOGLE_CLIENT_ID
ENV GOOGLE_CLIENT_ID=$GOOGLE_CLIENT_ID

ARG GOOGLE_CLIENT_SECRET
ENV GOOGLE_CLIENT_SECRET=$GOOGLE_CLIENT_SECRET

COPY . /JWT-Auth

WORKDIR /JWT-Auth

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

RUN make

WORKDIR /JWT-Auth/cmd/app

RUN go build .

CMD go run .
