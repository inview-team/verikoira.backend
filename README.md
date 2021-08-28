# Backend

### Build
```bash
$ make build
```

### Run
```bash
$ ./verikoira -config /path/to/config_file.json
```

### Config
```json
{
  "host": "0.0.0.0",
  "port": "1337",
  "logger":
  {
    "file": "api.log",
    "level": "debug"
  },
  "rabbitmq":
  {
    "address": "amqp://127.0.0.1:5672",
    "read_queue": "queue1",
    "write_queue": "queue2"
  }
}
```
