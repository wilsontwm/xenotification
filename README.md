### Simulate cron job on local

```
watch -n 60 curl --request POST http://localhost:7000/v1/cron/resend-notification
```

### Run unit test

```
make test
```
