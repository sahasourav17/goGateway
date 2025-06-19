# goGateway

![Go Version](https://img.shields.io/badge/go-1.23-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

A lightweight and scalable API Gateway in Go designed for microservice architectures. It offloads critical cross-cutting concerns by providing centralized JWT authentication, distributed tier based rate limiting with Redis, and resilience through circuit breakers, all configured dynamically via Consul.

---

## Core Features

- **Dynamic Routing:** Routes are loaded from Consul in real-time. Add, remove, or modify routes with no gateway restarts (hot-reloading).
- **Centralized Middleware:**
  - **JWT Authentication:** Secure routes with JWT validation.
  - **Tiered Rate Limiting:** Apply flexible, per-route, per-user-tier rate limits using the Sliding Window Log algorithm with Redis.
  - **Circuit Breaker:** Automatically detects failing downstream services and opens the circuit to prevent cascading failures.
- **Scalable Architecture:** Built on a stateless model for easy horizontal scaling.
- **Observability:** Structured JSON logging for easy parsing and analysis.
- **Containerized:** The entire environment (gateway, services, Consul, Redis) is managed via Docker Compose for easy, one-command setup.

## Getting Started

### Prerequisites

- Go (version 1.23+)
- Docker and Docker Compose

### Running the Project

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/sahasourav17/goGateway.git
    cd goGateway
    ```

2.  **Build and run the entire environment:**
    This single command starts the gateway, two mock services, Consul, and Redis.

    ```bash
    docker-compose up --build
    ```

    _(You can add the `-d` flag to run it in the background)_

3.  **Load the initial configuration into Consul:**
    In a new terminal window, run this command to upload the routing rules. The gateway will detect this and configure itself automatically.
    ```bash
    docker exec -i consul consul kv put gateway/config - < ./config/config.json
    ```

4.  **Access the Consul UI:** You can view the configuration and service health in your browser at `http://localhost:8500`.

## API Testing & Endpoints

### Generating a Test JWT

To access protected routes, you need a valid JWT.

1.  Go to **jwt.io**.
2.  Set the algorithm to **HS256**.
3.  Set the secret key to `a_very_insecure_default_secret`.
4.  Use one of the following payloads:
    - **Default User:** `{"user_id": "user-default-123"}`
    - **Premium User:** `{"user_id": "user-premium-456", "tier": "premium"}`
5.  Copy the generated token to use in the examples below.

### Example `curl` Commands

- **Test a Public Route (Rate Limit: 5 req/min)**

  ```bash
  curl -i http://localhost:8080/public/users/health
  ```

- **Test a Protected Route (No Token - Should Fail)**

  ```bash
  curl -i http://localhost:8080/api/users/profile
  # Expected: HTTP/1.1 401 Unauthorized
  ```

- **Test a Protected Route (With Token)**
  Replace `YOUR_JWT` with a token you generated.
  ```bash
  curl -i -H "Authorization: Bearer YOUR_JWT" http://localhost:8080/api/users/profile
  # Expected: HTTP/1.1 200 OK
  # Check the response headers for RateLimit-* details!
  ```

---
