# Technical White Paper: Unicorn Research Go
## Advanced Stochastic Modeling and Systematic Risk Analysis in a Concurrent Runtime
**Author:** Lukas Enderle  
**Version:** 2.1 (Expanded Multi-Strategy Focus)  
**Field:** Computational Finance / Backend Engineering  

---

### 1. Abstract
Unicorn Research Go is a high-throughput computational platform designed for the systematic stress testing of multi-asset portfolios. By leveraging a multi-threaded Monte Carlo simulation engine, the platform models macro-economic scenarios through correlated stochastic processes. This paper details the mathematical methodology, the statistical derivation of risk metrics, and the architectural implementation of both macro-level and sector-neutral hedge strategies within the Go runtime.

### 2. Methodology: Stochastic Modeling & Mathematical Foundation
The simulation engine models price dynamics using **Geometric Brownian Motion (GBM)**, a continuous-time stochastic process.

#### 2.1 The Stochastic Differential Equation (SDE)
Each asset $S_i$ is modeled according to the following SDE:
$$dS_t = \mu S_t dt + \sigma S_t dW_t$$
Where:
*   $\mu$ (Drift): The deterministic trend, calculated dynamically based on macro variables (Interest Rates and Oil Prices).
*   $\sigma$ (Volatility): Derived from historical 252-day log-return standard deviations.
*   $dW_t$: A Wiener process (Random Walk) representing the stochastic component.

#### 2.2 Correlated Multi-Asset Dynamics
To accurately model a portfolio, assets cannot be treated as independent variables. The system employs **Cholesky Decomposition** to transform independent Gaussian noise into correlated returns.
1.  **Correlation Matrix ($\Sigma$):** A $5 \times 5$ positive semi-definite matrix representing the historical relationships between Tech, Energy, Bonds, Crypto, and Gold.
2.  **Lower Triangular Matrix ($L$):** The system solves for $L$ such that $L L^T = \Sigma$.
3.  **Transformation:** Given a vector of independent random variables $Z$, the correlated vector $\epsilon$ is derived via $\epsilon = L Z$.

### 3. Sector Hedge Strategy: Factor Modeling & Option Convexity
Version 2.1 introduces a specialized module for analyzing **Sector-Neutral Long/Short strategies** augmented with derivative leverage.

#### 3.1 Real-Time Statistical Derivation
The platform transcends static modeling by integrating live market data for asset-specific calibration:
*   **Dynamic Beta Calculation**: The system automatically fetches historical daily candles for stocks (e.g., MU, NVDA) and benchmarks (e.g., QQQ, SPY). It computes the **Covariance($R_s, R_i$) / Variance($R_i$)** ratio over a 252-day trailing window to derive real-time Beta coefficients.
*   **Idiosyncratic Risk Extraction**: By applying the Factor Model identity ($\sigma_s^2 = \beta^2 \sigma_i^2 + \sigma_{idio}^2$), the engine isolates the residual volatility ($\sigma_{idio}$) unique to the specific company, allowing for precise "Alpha" simulation.

#### 3.2 Single-Factor Beta Model
Individual stock returns $R_i$ are modeled as a linear function of a Sector Index return $R_I$:
$$R_i = \beta_i R_I + \epsilon_i$$
Where $\beta_i$ is the sensitivity to the index and $\epsilon_i \sim N(0, \sigma_{idio}^2)$ represents the idiosyncratic risk specific to the company (Alpha source).

#### 3.3 Dynamic Derivative Pricing (Black-Scholes-Merton)
The strategy incorporates European Call Options, which are re-priced at every monthly step ($dt$) of the simulation to capture path-dependent effects:
$$C(S, t) = N(d_1)S_t - N(d_2)Ke^{-r(T-t)}$$
By embedding this formula within the Monte Carlo paths, the system accounts for:
*   **Theta Decay:** The non-linear loss of time value as the simulation approaches expiry.
*   **Gamma/Delta Dynamics:** The changing leverage of the position as the underlying index price fluctuates.

#### 3.3 Net Strategy P&L
The total strategy equity $V_{total}$ is calculated as the sum of the long stock values and the option value, minus the liability of the short index hedge:
$$V_{total}(t) = \sum_{i=1}^n \text{Stock}_i(t) + \text{Option}(t) - \text{ShortIndex}(t)$$

### 4. Quantitative Risk Analysis & Statistical Derivations
The platform focuses on the left-tail risk (worst-case outcomes) across 10,000+ paths.

#### 4.1 Risk Metrics (VaR & CVaR)
*   **VaR (99%):** Identified threshold loss that is not exceeded with a 99% probability.
*   **Expected Shortfall (ES):** Mean loss exceeding the VaR threshold, capturing the severity of "Tail Events."

#### 4.2 Strategy Performance Metrics
*   **Sharpe & Sortino Ratios:** Measures of risk-adjusted return, with the Sortino ratio specifically isolating downside volatility to better evaluate asymmetric option-based returns.
*   **Maximum Drawdown (MDD):** The average maximum peak-to-trough decline across all paths, quantifying capital impairment risk.
*   **Success Rate:** The probability $P(V_{final} > V_{initial})$, derived empirically from the simulation path set.

### 5. Software Engineering & High-Performance Concurrency
The implementation utilizes Go’s **CSP-style concurrency** (Communicating Sequential Processes):
*   **Worker Pools:** Partitioning simulation tasks across goroutines to scale linearly with CPU cores.
*   **Memory Optimization:** Avoidance of heap allocations in inner loops and localized PRNG (Pseudo-Random Number Generator) state to prevent global lock contention.

### 6. Conclusion
Unicorn Research Go provides a rigorous environment for the evaluation of complex financial strategies. By combining stochastic SDE modeling, factor-based correlation, and dynamic option pricing with a high-concurrency systems runtime, it delivers a professional-grade toolkit for modern quantitative research and risk management.
