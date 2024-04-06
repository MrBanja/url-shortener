# URL Shortener
Simple educational example of a URL shortener using Go, Redis and xid as a unique id generator.

## Setup
### Environment variables
You can skip this step and proceed to the next one if you want to use the default values from `.env.example`.

The following environment variables are required:
- `REDIS_PASSWORD` — Redis password.
- `REDIS_CONN_STR` — Redis connection string.
- `PORT` — Service port.
- `ADDR` — Full address of the service (including port above).

### Run
1. `make prepare`. _Optional step_. Will create a `.env` file with the default values from `.env.example`.
2. `make up`. Will spin up a docker compose.

To **stop** the docker compose run `make stop`.

## Usage
_Assuming the service is running on `http://localhost:8080`._

### Shorten URL
```bash
$ curl -X POST http://localhost:8080/encode -d '{"url": "https://google.com"}'

co8j2j1ggvkc73enu60g
```
### Redirect
```bash
$ curl -X GET http://localhost:8080/decode\?short\=co8j2j1ggvkc73enu60g 

<a href="https://google.com">Found</a>.
```

For more information, please refer to the [OpenAPI Doc](doc/openapi.yaml).

## Q&A
### Describe your solution. What tradeoffs did you make while designing it, and why?

This is a simplified yet functional and scalable URL shortener. My main goal was to create a solution that is effective, straightforward, and requires minimal dependencies for production readiness.

I opted to use Redis as the storage backend due to its speed, simplicity, and scalability. Redis excels in both read performance and persistence, making it a suitable choice for this use case.

The URL shortening algorithm was a key consideration. There are various methods to generate unique IDs, and I selected a method that generates unique IDs without specific constraints. For this purpose, I employed the xid package, known for its speed, uniqueness, and compactness compared to UUIDs.

Tradeoffs:
- To maintain simplicity, I did not implement an expiration time for the short links.
- The resulting short URL might be longer due to the chosen method of generating unique IDs, which is the primary tradeoff.

### If this were a real project, how would you improve it further?
- Implement expiration times.
- Add a caching layer (e.g., Bloom Filter) to reduce Redis requests and optimize performance.
- Introduce rate limiting to control access and prevent abuse.
- Modify the URL generation process to potentially use base64 encoding of UID pools if URL length becomes problematic.
- Consider adding a tracking system to monitor usage and engagement, as such services are often used for analytics.

### What is the math & logic behind the "short URL" if there is any?
The core concept involves generating a unique identifier for each URL. While various methods exist, the chosen approach typically uses unique ID generators. However, this approach may result in longer IDs.

A more efficient technique involves using base64 encoding of unique IDs. This method is preferred for shorter UID length, where a large pool of unique IDs must be efficiently managed and distributed.

### Scaling
This service is designed for horizontal scalability. Instances do not share state, except for Redis, which is a distributed database capable of seamless scaling.