# LOGLITE

I will try and create a simple, lightweight, and easy to setup (and use) logging system

## Configuration

LogLite allows you to configure its behavior using a YAML configuration file.

## Example `config.yaml`

```yaml
debug_level: 'DEBUG'
protocol: 'HTTP'
port: 8080
log_file: 'logs/app.log'
max_connections: 50
```

## How to run?

```bash
air
```

and it runs a http server on: http://localhost:8080/

It is fairly simple. If you want to manually build and run then:

```bash
go build .
./LogLite
```

Also pretty simple
