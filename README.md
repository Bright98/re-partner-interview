# Pack Calculator

Calculates the optimal number of packs to ship for a given order, following these rules (in priority order):

1. Only whole packs can be sent — packs cannot be broken open.
2. Send the fewest **items** possible to cover the order.
3. Within the same item total, send the fewest **packs** possible.

Available pack sizes: **250 · 500 · 1,000 · 2,000 · 5,000**

## Requirements

- Go 1.21+

## Run

```bash
go run .
```

Then open **http://localhost:8080/ui/index.html** in your browser.

## API

```
GET /packs?order=<n>
```

**Example**

```bash
curl 'http://localhost:8080/packs?order=12001'
```

```json
{
  "order": 12001,
  "packs": [
    { "size": 5000, "count": 2 },
    { "size": 2000, "count": 1 },
    { "size": 250,  "count": 1 }
  ],
  "total_items": 12250,
  "total_packs": 4
}
```

**Errors** return HTTP 400 with `{ "error": "..." }`.

## Tests

```bash
go test ./packs/...
```

## Project Structure

```
.
├── main.go          # HTTP server
├── packs/
│   ├── packs.go     # core logic
│   └── packs_test.go
└── ui/
    └── index.html   # browser UI (embedded into the binary)
```
