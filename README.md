```shell
curl -X POST http://localhost:8080/fingerprint \
  -H "Content-Type: application/json" \
  -d '{"visitor_id": "123", "user_agent": "Mozilla/5.0", "components": {"key": "value"}, "dpr": "3.0"}'
```

```shell
     touch /tmp/fingerprints.db
     docker run --name fp-analytics --network="host" -v /tmp/fingerprints.db:/app/fingerprints.db -d fp-analytics
```