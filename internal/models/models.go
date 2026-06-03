package models

import "time"

type FuturePricingRequest struct {
	SpotPrice     float64 `json:"spot_price"`
	RiskFreeRate  float64 `json:"risk_free_rate"`
	DividendYield float64 `json:"dividend_yield"`
	TimeToExpiry  float64 `json:"time_to_expiry_years"`
}

type FuturePricingResponse struct {
	FuturePrice float64 `json:"future_price"`
	Basis       float64 `json:"basis"`
}

type SwapPricingRequest struct {
	Notional     float64 `json:"notional"`
	FixedRate    float64 `json:"fixed_rate"`
	FloatingRate float64 `json:"floating_rate"`
	TenorYears   float64 `json:"tenor_years"`
	Frequency    int     `json:"frequency"`
}

type SwapPricingResponse struct {
	NPV           float64 `json:"npv"`
	FairSwapRate  float64 `json:"fair_swap_rate"`
	FixedLegValue float64 `json:"fixed_leg_value"`
}

type OptionGreeks struct {
	Delta float64 `json:"delta"`
	Gamma float64 `json:"gamma"`
	Theta float64 `json:"theta"`
	Vega  float64 `json:"vega"`
	Rho   float64 `json:"rho"`
}

type SimulationRequest struct {
	InterestRate  float64 `json:"interest_rate"`
	OilPrice      float64 `json:"oil_price"`
	PosATech      float64 `json:"pos_a_tech"`
	PosBEnergy    float64 `json:"pos_b_energy"`
	PosCBonds     float64 `json:"pos_c_bonds"`
	PosDCrypto    float64 `json:"pos_d_crypto"`
	PosEGold      float64 `json:"pos_e_gold"`
	VolATech      float64 `json:"vol_a_tech"`
	VolBEnergy    float64 `json:"vol_b_energy"`
	VolCBonds     float64 `json:"vol_c_bonds"`
	VolDCrypto    float64 `json:"vol_d_crypto"`
	VolEGold      float64 `json:"vol_e_gold"`
	DurationYears int     `json:"duration_years"`
	NumPaths      int     `json:"num_paths"`
}

type SimulationResponse struct {
	Status       string       `json:"status"`
	InputSummary InputSummary `json:"input_summary"`
	Metrics      Metrics      `json:"metrics"`
	ChartData    [][]float64  `json:"chart_data"`
}

type SectorHedgeRequest struct {
	IndexSymbol    string    `json:"index_symbol"`
	IndexVol       float64   `json:"index_vol"`
	StockSymbols   []string  `json:"stock_symbols"`
	StockPositions []float64 `json:"stock_positions"`
	StockBetas     []float64 `json:"stock_betas"`
	StockVols      []float64 `json:"stock_vols"`
	OptionStrike   float64   `json:"option_strike"`
	OptionExpiry   float64   `json:"option_expiry_years"`
	OptionQuantity float64   `json:"option_quantity"`
	ShortIndexPos  float64   `json:"short_index_pos"`
	DurationYears  float64   `json:"duration_years"`
	RiskFreeRate   float64   `json:"risk_free_rate"`
	NumPaths       int       `json:"num_paths"`
}

type SectorHedgeResponse struct {
	Status    string       `json:"status"`
	NumPaths  int          `json:"num_paths"`
	Metrics   HedgeMetrics `json:"metrics"`
	ChartData [][]float64  `json:"chart_data"`
	Greeks    OptionGreeks `json:"greeks"`
}

type HedgeMetrics struct {
	FinalValueMean     float64 `json:"final_value_mean"`
	WinRate            float64 `json:"win_rate"`
	MaxDrawdown        float64 `json:"max_drawdown"`
	AlphaContribution  float64 `json:"alpha_contribution"`
	HedgeEffectiveness float64 `json:"hedge_effectiveness"`
}

type BacktestRequest struct {
	DurationYears int                `json:"duration_years"`
	Weights       map[string]float64 `json:"weights"`
}

type BacktestResponse struct {
	Dates       []string  `json:"dates"`
	EquityCurve []float64 `json:"equity_curve"`
	TotalReturn float64   `json:"total_return"`
	MaxDrawdown float64   `json:"max_drawdown"`
	SharpeRatio float64   `json:"sharpe_ratio"`
	Volatility  float64   `json:"volatility"`
}

type InputSummary struct {
	TotalInvestment  float64          `json:"total_investment"`
	CalculatedDrifts CalculatedDrifts `json:"calculated_drifts"`
}

type CalculatedDrifts struct {
	Tech   float64 `json:"tech"`
	Energy float64 `json:"energy"`
	Bonds  float64 `json:"bonds"`
	Crypto float64 `json:"crypto"`
	Gold   float64 `json:"gold"`
}

type Metrics struct {
	Var99Percent      float64 `json:"var_99_percent"`
	ExpectedShortfall float64 `json:"expected_shortfall_99_percent"`
	ExpectedMeanEur   float64 `json:"expected_mean_eur"`
	SharpeRatio       float64 `json:"sharpe_ratio"`
	SortinoRatio      float64 `json:"sortino_ratio"`
	MaxDrawdown       float64 `json:"max_drawdown"`
}

type Asset struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Symbol string `json:"symbol"`
	Type   string `json:"type"`
	Name   string `json:"name"`
}

type PortfolioAsset struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	AssetID       uint    `json:"asset_id"`
	Asset         Asset   `gorm:"foreignKey:AssetID" json:"asset"`
	Quantity      float64 `json:"quantity"`
	PurchasePrice float64 `json:"purchase_price"`
}

type SimulationResult struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time `json:"timestamp"`
	InputJSON       string    `json:"input"`
	TotalInvestment float64   `json:"total_investment"`
	Var99           float64   `json:"var_99"`
	ES99            float64   `json:"es_99"`
	MeanEur         float64   `json:"mean_eur"`
	SharpeRatio     float64   `json:"sharpe_ratio"`
	SortinoRatio    float64   `json:"sortino_ratio"`
	MaxDrawdown     float64   `json:"max_drawdown"`
}

type SectorHedgeResult struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time `json:"timestamp"`
	InputJSON      string    `json:"input"`
	FinalValueMean float64   `json:"final_value_mean"`
	WinRate        float64   `json:"win_rate"`
	MaxDrawdown    float64   `json:"max_drawdown"`
}

type Dashboard struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `json:"name"`
	InterestRate  float64   `json:"interest_rate"`
	OilPrice      float64   `json:"oil_price"`
	PosATech      float64   `json:"pos_a_tech"`
	PosBEnergy    float64   `json:"pos_b_energy"`
	PosCBonds     float64   `json:"pos_c_bonds"`
	PosDCrypto    float64   `json:"pos_d_crypto"`
	PosEGold      float64   `json:"pos_e_gold"`
	DurationYears int       `json:"duration_years"`
	NumPaths      int       `json:"num_paths"`
	CreatedAt     time.Time `json:"created_at"`
}

type SectorHedgeDashboard struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	IndexSymbol string    `json:"index_symbol"`
	CreatedAt   time.Time `json:"created_at"`
}

type PortfolioRiskReport struct {
	TotalValue       float64     `json:"total_value"`
	PortfolioMetrics Metrics     `json:"portfolio_metrics"`
	AssetBreakdown   []AssetRisk `json:"asset_breakdown"`
	Timestamp        time.Time   `json:"timestamp"`
}

type AssetRisk struct {
	Symbol       string  `json:"symbol"`
	Value        float64 `json:"value"`
	Weight       float64 `json:"weight"`
	Volatility   float64 `json:"volatility"`
	Contribution float64 `json:"risk_contribution"`
}

type StressTestRequest struct {
	ScenarioName string             `json:"scenario_name"`
	Weights      map[string]float64 `json:"weights"`
}

type StressTestResponse struct {
	ScenarioName  string        `json:"scenario_name"`
	DescriptionEN string        `json:"description_en"`
	DescriptionDE string        `json:"description_de"`
	TotalDrop     float64       `json:"total_drop"`
	AssetImpacts  []AssetImpact `json:"asset_impacts"`
}

type AssetImpact struct {
	Asset       string  `json:"asset"`
	Weight      float64 `json:"weight"`
	Shock       float64 `json:"shock"`
	ValueImpact float64 `json:"value_impact"`
}

type CustomSimulationRequest struct {
	Symbols       []string  `json:"symbols"`
	Allocations   []float64 `json:"allocations"`
	DurationYears int       `json:"duration_years"`
	NumPaths      int       `json:"num_paths"`
	RiskFreeRate  float64   `json:"risk_free_rate"`
}

type CustomSimulationResponse struct {
	Status    string      `json:"status"`
	NumPaths  int         `json:"num_paths"`
	Metrics   Metrics     `json:"metrics"`
	ChartData [][]float64 `json:"chart_data"`
}
