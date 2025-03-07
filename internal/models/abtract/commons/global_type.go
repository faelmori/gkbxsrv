package commons

type PingImpl struct {
	Ping string `json:"ping"`
}

type TableHandler struct {
	Rows map[int]map[string]string
}

func (h *TableHandler) GetHeaders() []string {
	var headers []string
	for _, row := range h.Rows {
		for k := range row {
			headers = append(headers, k)
		}
		break
	}
	return headers
}
func (h *TableHandler) GetRows() [][]string {
	var rows [][]string
	for _, row := range h.Rows {
		var r []string
		for _, v := range row {
			r = append(r, v)
		}
		rows = append(rows, r)
	}
	return rows
}
