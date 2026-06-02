package models

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
	Var99Percent     float64 `json:"var_99_percent"`
	ExpectedShortfall float64 `json:"expected_shortfall_99_percent"`
	ExpectedMeanEur  float64 `json:"expected_mean_eur"`
}
