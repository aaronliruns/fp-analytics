
### Initializing the database
```shell
     touch /tmp/fingerprints.db
     docker run --name fp-analytics --network="host" -v /tmp/fingerprints.db:/app/fingerprints.db -d fp-analytics
```

### Collecting the fingerprints
```shell
# Store a fingerprint
curl -X POST http://localhost:8080/fingerprint \
  -H "Content-Type: application/json" \
  -d '{"visitor_id": "123", "user_agent": "Mozilla/5.0", "components": "{\"key\":\"value\"}", "dpr": "3.0"}'
```
On success, returns HTTP status 201 Created with empty response body.

Possible error responses:
```json
{"error": "Invalid request payload"}
{"error": "Failed to save fingerprint"}
{"error": "Invalid value for field VisitorID"}
```

### Querying fingerprint data

#### Get total count of fingerprints
```shell
curl http://localhost:8080/fingerprints/count
```
Example response:
```json
{
    "count": 42
}
```

#### Get specific fingerprint by row number
```shell
curl http://localhost:8080/fingerprints/row?row=5
```
Example response:
```json
{
    "row_number": 5,
    "user_agent": "Mozilla/5.0 (X11; Linux x86_64)",
    "screen_resolution": [1920, 1080],
    "hardware_concurrency": 8,
    "platform": "Linux x86_64",
    "touch_support": {
        "maxTouchPoints": 5,
        "touchEvent": true,
        "touchStart": true
    },
    "video_card": {
        "vendor": "NVIDIA",
        "model": "GeForce GTX 1060"
    },
    "architecture": 255,
    "dpr": 2.0
}
```

Possible error responses:
```json
{"error": "Row number is required"}
{"error": "Invalid row number"}
{"error": "Row not found"}
{"error": "Failed to get fingerprint row"}
{"error": "Failed to parse components data"}
```