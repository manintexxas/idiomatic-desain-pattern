package app

func InitDefaultData() {
	if len(EmployeeDict) == 0 {
		NewEmployee("Budi Santoso")
		NewEmployee("Siti Rahma")
	}

	if len(CustDict) == 0 {
		NewCustomer("Andi Wijaya", true)
		NewCustomer("Dewi Lestari", false)

		custOpname := ID("CST-0000")
		CustDict[custOpname] = Customer{
			CustID:   custOpname,
			CustName: "OPNAME",
			IsMember: false,
		}
	}

	if len(InventoryDict) == 0 {
		NewInventory("2026-12-31 00:00:00", "Lada Bubuk", 50, ItemActive)
		NewInventory("2026-12-31 00:00:00", "Kopi Hitam", 100, ItemActive)
		NewInventory("2026-12-31 00:00:00", "Teh Hijau", 75, ItemActive)
		NewInventory("2026-12-31 00:00:00", "Roti Tawar", 30, ItemActive|ItemRestricted)
		NewInventory("2026-12-31 00:00:00", "Susu UHT", 45, ItemActive)
	}
}
