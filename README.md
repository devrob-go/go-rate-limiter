# Go Rate Limiter

A high-performance, modular, and scalable rate-limiting library written in Go. This package supports both in-memory and Redis backends, allowing developers to throttle requests efficiently in both single-node and distributed systems.

---

## **Features**
- **In-Memory Backend**: For single-node applications with low latency requirements.
- **Redis Backend**: Distributed rate-limiting with persistence and scalability.
- **Thread-Safe**: Optimized with `sync.Map` and safe for concurrent operations.
- **Customizable**: Define custom refill rates and limits per key.
- **Unit Tests and Benchmarks**: Ensures reliability and performance under load.

---

## **Installation**
```bash
go get github.com/devrob-go/go-rate-limiter
