more detailed# White Paper: Unicorn Research Go
## A High-Performance Macro Quant Stress Testing Platform
**Author:** Lukas Enderle  
**Date:** June 2, 2026  
**Stack:** Go 1.22+, Gin, GORM, SQLite, HTMX, Tailwind CSS, Chart.js

---

### 1. Executive Summary
Unicorn Research Go is a high-performance quantitative finance backend and interactive simulation dashboard. It enables financial analysts to perform macro-level stress testing on multi-asset portfolios using Monte Carlo simulations. The platform is designed with a focus on concurrency, data isolation, and modern web patterns (HTMX/REST), providing both a headless API for programmatic access and a rich UI for manual scenario research.

### 2. Core Quantitative Engine
The heart of the application is a multi-threaded Monte Carlo simulation engine implemented in pure Go.

*   **Concurrency Model:** Utilizing Go's goroutines and `sync.WaitGroup`, the simulator distributes tens of thousands of paths across all available CPU cores. This allows for near-instantaneous results even for complex, multi-year projections.
*   **Stochastic Modeling:** The engine simulates asset price movements using Geometric Brownian Motion (GBM). 
*   **Cholesky Decomposition:** To handle multi-asset portfolios, the system incorporates a Cholesky decomposition algorithm to simulate correlated returns across five distinct asset classes (Tech, Energy, Bonds, Crypto, and Gold).
*   **Historical Volatility:** Instead of hardcoded values, the system integrates with the MarketData API to calculate real-world annualized volatility based on 252-day log-return standard deviations.

### 3. Advanced Risk Metrics
The platform provides a comprehensive suite of "Quant-First" metrics, moving beyond simple mean returns:
*   **Value at Risk (VaR 99%):** Quantifies the maximum expected loss over a specific timeframe with 99% confidence.
*   **Expected Shortfall (CVaR):** Calculates the average loss in the most extreme 1% of scenarios.
*   **Risk-Adjusted Returns:** Implements both **Sharpe Ratio** (total risk) and **Sortino Ratio** (downside risk) to evaluate investment quality.
*   **Maximum Drawdown:** Tracks the average peak-to-trough decline across all paths to evaluate potential "investor pain."

### 4. Technical Architecture
The project follows a Clean Architecture pattern, ensuring maintainability and scalability:

*   **Headless REST API:** A robust set of JSON endpoints allows external frontends to authenticate, run simulations, and manage portfolios.
*   **JWT & RBAC:** Implements industry-standard JSON Web Tokens for authentication and a flexible Role-Based Access Control system (User vs. Admin).
*   **HTMX Integration:** For the internal dashboard, the project uses HTMX to provide a "Single Page Application" feel with zero-refresh updates, powered entirely by server-side Go templates.
*   **Database Isolation:** Using GORM with SQLite, the system enforces strict data isolation at the query level, ensuring users only access their own scenarios and dashboards.
*   **Bootstrap Security:** Features an automated admin bootstrapping process with enforced password rotation policies on first login.

### 5. Deployment & Scalability
*   **Dockerized:** The application is containerized, ensuring consistent environments from development to production.
*   **Observability:** Integrated audit logging (User Management logs) and system error logging provide high transparency for administrators.
*   **Branching Strategy:** The project maintains distinct architectural paths (HTMX vs. Pure REST) via Git branches, demonstrating versatility in delivery methods.

---

### Conclusion
Unicorn Research Go demonstrates a deep understanding of both high-performance backend engineering and quantitative financial theory. By leveraging Go's native strengths in concurrency and type safety, it delivers a professional-grade tool capable of assisting in complex macro-financial decision-making.
