```shell
   curl -X POST http://localhost:8080/fingerprint \
   -H "Content-Type: application/json" \
   -d '{ 
      "visitor_id": "test", 
      "user_agent": "Mozilla/5.0", 
      "components": "{\"key\":\"value\"}"
   }'
```