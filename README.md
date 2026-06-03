# Unicorn Research Go (Public Edition)

A Go implementation of the Macro Quant Stress Testing API and UI, focusing on speed and accessibility without user management overhead.

## Features

- **Interactive UI**: Built with [HTMX](https://htmx.org/) and Tailwind CSS.
- **Monte Carlo Simulation**: Multi-threaded portfolio stress testing.
- **Public REST API**: Full JSON access to simulations and dashboards.
- **Scenario Management**: Save and load simulation parameters as named scenarios.
- **Advanced Metrics**: Sharpe Ratio, Sortino Ratio, and Max Drawdown calculation.
- **Historical Volatility**: Automatic fetching of market data proxies.

## Getting Started

### Prerequisites

- Go 1.22+

### Running the Application

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`.

## API Endpoints

### POST /api/v1/simulate
Runs a macro stress test simulation.

### GET /api/v1/risk-report
Generates a risk report based on the global portfolio.

### GET /api/v1/dashboards
Lists all saved scenarios.

## License

MIT
