FROM       golang:1.16
WORKDIR    /verikoira
COPY       go.mod .
COPY       go.sum .
RUN        go mod download
COPY       . .
RUN        go build -o verikoira ./cmd/main.go
RUN ls
ENTRYPOINT ["./verikoira"]
