package models

type PingImpl struct {
	Ping string `json:"ping"`
}

type RegisterUserInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
type RegisterUserWithEmailInput struct {
	Username string `json:"username" binding:"omitempty"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
type RegisterUserWithUsernameInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"omitempty"`
	Password string `json:"password" binding:"required,min=8"`
}
type LoginWithEmailInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
type LoginWithUsernameInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}
type RegisterProductInput struct {
	Name     string  `json:"name" binding:"required"`
	Price    float64 `json:"price" binding:"required"`
	Depart   string  `json:"depart" binding:"omitempty"`
	Category string  `json:"category" binding:"omitempty"`
	Cost     float64 `json:"cost" binding:"default=0"`
	Stock    int     `json:"stock" binding:"default=0"`
	Reserve  int     `json:"reserve" binding:"default=0"`
	Balance  int     `json:"balance" binding:"default=0"`
	Synced   bool    `json:"synced" binding:"default=false"`
	LastSync string  `json:"last_sync" binding:"default=current_time"`
}
type UpdateProductInput struct {
	Name     string  `json:"name" binding:"omitempty"`
	Price    float64 `json:"price" binding:"omitempty"`
	Depart   string  `json:"depart" binding:"omitempty"`
	Category string  `json:"category" binding:"omitempty"`
	Cost     float64 `json:"cost" binding:"omitempty"`
	Stock    int     `json:"stock" binding:"omitempty"`
	Reserve  int     `json:"reserve" binding:"omitempty"`
	Balance  int     `json:"balance" binding:"omitempty"`
	Synced   bool    `json:"synced" binding:"omitempty"`
	LastSync string  `json:"last_sync" binding:"omitempty"`
}

type TableHandler struct {
	rows map[int]map[string]string
}

func (h *TableHandler) GetArrayMap() map[string][]string {
	var arrayMap = make(map[string][]string)
	for _, row := range h.rows {
		for k, v := range row {
			arrayMap[k] = append(arrayMap[k], v)
		}
	}
	return arrayMap
}
func (h *TableHandler) GetHashMap() map[string]string {
	var hashMap = make(map[string]string)
	for _, row := range h.rows {
		for k, v := range row {
			hashMap[k] = v
		}
	}
	return hashMap
}
func (h *TableHandler) GetObjectMap() []map[string]string {
	var objectMap []map[string]string
	for _, row := range h.rows {
		var obj = make(map[string]string)
		for k, v := range row {
			obj[k] = v
		}
		objectMap = append(objectMap, obj)
	}
	return objectMap
}
func (h *TableHandler) GetByteMap() map[string][]byte {
	var byteMap = make(map[string][]byte)
	for _, row := range h.rows {
		for k, v := range row {
			byteMap[k] = []byte(v)
		}
	}
	return byteMap
}
func (h *TableHandler) GetHeaders() []string {
	var headers []string
	for _, row := range h.rows {
		for k := range row {
			headers = append(headers, k)
		}
		break
	}
	return headers
}
func (h *TableHandler) GetRows() [][]string {
	var rows [][]string
	for _, row := range h.rows {
		var r []string
		for _, v := range row {
			r = append(r, v)
		}
		rows = append(rows, r)
	}
	return rows
}
