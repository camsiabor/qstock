package dict

type StockInfo struct {
	Name              string
	OpenPrice         string
	ClosePrice        string
	NowPrice          string
	TodayMax          string
	TodayMin          string
	TradeNum          string
	TraceAmount       string
	Buy1_n            string
	Buy1_m            string
	Buy2_n            string
	Buy2_m            string
	Buy3_n            string
	Buy3_m            string
	Buy4_n            string
	Buy4_m            string
	Buy5_n            string
	Buy6_m            string
	Sell1_n           string
	Sell1_m           string
	Sell2_n           string
	Sell2_m           string
	Sell3_n           string
	Sell3_m           string
	Sell4_n           string
	Sell4_m           string
	Sell5_n           string
	Sell5_m           string
	Date              string
	Time              string
	Diff_money        string
	Diff_rate         string
	Swing             string
	Turnover          string
	Pe                string
	Pb                string
	Highlimit         string
	Downlimit         string
	All_value         string
	Circulation_value string
	Currcapital       string
	Totalcapital      string
	Max52             string
	Min52             string
	AppointRate       string
	AppointDiff       string
	State             int
}

type ShowResponseBody struct {
	Ret_code int
	List     []StockInfo
}

type ShowResponse struct {
	Showapi_res_error string
	Showapi_res_id    string
	Showapi_res_code  int
	Showapi_res_body  ShowResponseBody
}
