package main

import (
	"encoding/json"
	"log"
	"net/http"

	"dict/app"
)

func main() {
	app.InitDefaultData()

	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("../wwwroot")))

	mux.HandleFunc("POST /inventory", func(w http.ResponseWriter, r *http.Request) {
		var req app.CreateInventoryRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		inv, err := app.NewInventory(req.Date, req.ItemName, req.Stock, app.Flag(req.Flag))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(inv)
	})

	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		responseMap := make(map[app.ID]map[string]any)
		for id, inv := range app.InventoryDict {
			responseMap[id] = map[string]any{
				"ItemExp":  app.GetTimestamp(inv.ItemExp),
				"ItemID":   inv.ItemID,
				"ItemName": inv.ItemName,
				"Stock":    inv.Stock,
				"ItemFlag": inv.ItemFlag,
			}
		}

		json.NewEncoder(w).Encode(responseMap)
	})

	mux.HandleFunc("GET /inventory/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := app.ID(r.PathValue("id"))
		inv, exists := app.InventoryDict[id]
		if !exists {
			http.Error(w, "ID not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]any{
			"ItemExp":  app.GetTimestamp(inv.ItemExp),
			"ItemID":   inv.ItemID,
			"ItemName": inv.ItemName,
			"Stock":    inv.Stock,
			"ItemFlag": inv.ItemFlag,
		}

		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("PUT /inventory/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := app.ID(r.PathValue("id"))
		inv, exists := app.InventoryDict[id]
		if !exists {
			http.Error(w, "ID not found", http.StatusNotFound)
			return
		}

		var req app.CreateInventoryRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cleanedName, err := app.NameFormat(req.ItemName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		inv.ItemName = cleanedName
		inv.Stock = req.Stock
		inv.ItemFlag = app.Flag(req.Flag)
		app.InventoryDict[id] = inv

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(inv)
	})

	mux.HandleFunc("DELETE /inventory/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := app.ID(r.PathValue("id"))
		if err := app.DeleteData(app.InventoryDict, id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("POST /employee", func(w http.ResponseWriter, r *http.Request) {
		var req app.CreateEmployeeRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		emp, err := app.NewEmployee(req.EmpName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(emp)
	})

	mux.HandleFunc("GET /employee", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(app.EmployeeDict)
	})

	mux.HandleFunc("DELETE /employee/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := app.ID(r.PathValue("id"))
		if err := app.DeleteData(app.EmployeeDict, id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("POST /customer", func(w http.ResponseWriter, r *http.Request) {
		var req app.CreateCustomerRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cust, err := app.NewCustomer(req.CustName, req.IsMember)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(cust)
	})

	mux.HandleFunc("GET /customer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(app.CustDict)
	})

	mux.HandleFunc("DELETE /customer/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := app.ID(r.PathValue("id"))
		if err := app.DeleteData(app.CustDict, id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("POST /inventory/{id}/add", func(w http.ResponseWriter, r *http.Request) {
		id := app.ID(r.PathValue("id"))
		inv, exists := app.InventoryDict[id]
		if !exists {
			http.Error(w, "Inventory ID not found", http.StatusNotFound)
			return
		}

		var req app.TransactionRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		trx, err := inv.AddStock(app.ID(req.ReceiverID), req.Amount, req.Desc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		app.InventoryDict[id] = inv

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(trx)
	})

	mux.HandleFunc("POST /inventory/{id}/reduce", func(w http.ResponseWriter, r *http.Request) {
		id := app.ID(r.PathValue("id"))
		inv, exists := app.InventoryDict[id]
		if !exists {
			http.Error(w, "Inventory ID not found", http.StatusNotFound)
			return
		}

		var req app.TransactionRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		trx, err := inv.ReduceStock(app.ID(req.EmpID), app.ID(req.CustID), req.Amount, req.Desc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		app.InventoryDict[id] = inv

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(trx)
	})

	mux.HandleFunc("GET /transactions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Konversi Timestamp int64 pada setiap Transaction menjadi format string yang mudah dibaca
		responseMap := make(map[app.ID]map[string]any)
		for id, trx := range app.TransactionsDict {
			responseMap[id] = map[string]any{
				"TransID":   trx.TransID,
				"ItemID":    trx.ItemID,
				"EmpID":     trx.EmpID,
				"CustID":    trx.CustID,
				"Amount":    trx.Amount,
				"Timestamp": app.GetTimestamp(trx.Timestamp),
				"Desc":      trx.Desc,
			}
		}

		json.NewEncoder(w).Encode(responseMap)
	})

	port := ":8080"
	log.Printf("Server berjalan di http://localhost%s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server gagal dijalankan: %v", err)
	}
}
