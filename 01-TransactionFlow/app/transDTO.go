package app

type InventoryResponseDTO struct {
	ItemID   string `json:"ItemId"`
	ItemName string `json:"itemName"`
	Stock    uint8  `json:"stock"`
	Status   string `json:"status"`
	ItemExp  string `json:"itemExp"`
}

type EmployeeResponseDTO struct {
	EmpID   string `json:"empId"`
	EmpName string `json:"empName"`
}

type CustomerResponseDTO struct {
	CustID   string `json:"custId"`
	CustName string `json:"custName"`
	IsMember bool   `json:"isMember"`
}

type TransactionResponseDTO struct {
	TransID   string `json:"transId"`
	Timestamp string `json:"timestamp"`
	ItemID    string `json:"itemId"`
	EmpID     string `json:"empId"`
	CustID    string `json:"custId"`
	Desc      string `json:"desc"`
	Amount    uint8  `json:"amount"`
	CurStock  uint8  `json:"curStock"`
	NowStock  uint8  `json:"nowStock"`
}

type CreateInventoryRequestDTO struct {
	Date     string `json:"date"`
	ItemName string `json:"itemName"`
	Stock    uint8  `json:"stock"`
	Flag     uint8  `json:"flag"`
}

type CreateEmployeeRequestDTO struct {
	EmpName string `json:"empName"`
}

type CreateCustomerRequestDTO struct {
	CustName string `json:"custName"`
	IsMember bool   `json:"isMember"`
}

type TransactionRequestDTO struct {
	ReceiverID string `json:"receiverId,omitempty"`
	EmpID      string `json:"empId,omitempty"`
	CustID     string `json:"custId,omitempty"`
	Amount     uint8  `json:"amount"`
	Desc       string `json:"desc"`
}
