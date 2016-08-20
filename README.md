# A simple reverse proxy

## Installation

    # compile server and client
    go get -v github.com/Dreamacro/singular/cmd/singulard
    go get -v github.com/Dreamacro/singular/cmd/singular

## Getting started

### Server

    # run server
    singulard -port 8000

### Options

* `-port port`, e.g., `-port 8000`, default `8080`
* `-log logPath`, e.g., `-log proxy.log`, default log is tty mode
* `-tls`, with tls
* `-cert cert`, e.g., `-cert cert.pem`, default `cert.pem`
* `-key key`, e.g., `-key cert.key`, default `cert.key`

### Client

	# run client
    singular -config config.yml

### Options

* `-config yaml`, e.g., `-config config.yaml`, default `config.yml`

```yaml
server_addr: domain.com:8080
proxy:
    ssh: tcp://0.0.0.0:22
    docker: unix:///var/run/docker.sock
```    

* `-log logPath`, e.g., `-log proxy.log`, default log is tty mode
* `-tls`, with tls
* `-cert cert`, e.g., `-cert cert.pem`, default `cert.pem`
* `-key key`, e.g., `-key cert.key`, default `cert.key`

## Start with tls
    singulard -tls
    singular -tls
