# Unicorn Research Go

A Go implementation of the Macro Quant Stress Testing API, inspired by the Python `unicorn_research` project.

## Features

- REST API built with [Gin](https://github.com/gin-gonic/gin)
- Monte Carlo simulation for portfolio stress testing (Simplified implementation)
- CORS support for frontend integration
- Clean architecture with separate models, handlers, and simulation logic

## Project Structure

- `cmd/server/`: Entry point for the Go server.
- `internal/api/`: REST API handlers and routing.
- `internal/models/`: Shared data models for requests and responses.
- `internal/simulation/`: Core simulation logic.

## Getting Started

### Prerequisites

- Go 1.22+

### Installation

1. Clone the repository (or copy the files).
2. Install dependencies:
   ```bash
   go mod tidy
   ```

### Running the Server

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`.

### API Endpoints

#### POST /api/v1/simulate

Runs a macro stress test simulation based on the provided parameters.

**Request Body:**
```json
{
  "interest_rate": 4.25,
  "oil_price": 82.50,
  "pos_a_tech": 307.70,
  "pos_b_energy": 100.00,
  "pos_c_bonds": 200.00,
  "vol_a_tech": 0.18,
  "vol_b_energy": 0.18,
  "vol_c_bonds": 0.05,
  "duration_years": 5,
  "num_paths": 1000
}
```

**Response Body:**
```json
{
  "status": "success",
  "input_summary": {
    "total_investment": 607.7,
    "calculated_drifts": {
      "tech": 5.63,
      "energy": 7.68,
      "bonds": 4.25
    }
  },
  "metrics": {
    "var_99_percent": 1.23,
    "expected_mean_eur": 745.21
  },
  "chart_data": [...]
}
```

## License

MIT
