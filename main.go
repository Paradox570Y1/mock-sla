// mock_sla is a lightweight mock server that simulates the external
// routing/raise-complaint API for local testing.
//
// Usage:
//
//	go run ./cmd/mock_sla/main.go
//
// Then set in .env:
//
//	SLA_BASE_URL=http://localhost:9293
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type request struct {
	SubOrderID    string `json:"subOrderId"`
	VendorID      string `json:"vendorId"`
	Message       string `json:"message"`
	RaisedByEmail string `json:"raisedByEmail"`
}

func main() {
	// 1. Endpoint: raise-complaint
	raiseComplaintHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		apiKey := r.Header.Get("X-Api-Key")
		channel := r.Header.Get("X-Routing-Channel")

		log.Printf("📥 Received raise-complaint | X-Api-Key=%s | X-Routing-Channel=%s", apiKey, channel)

		if apiKey != "ERP" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "INVALID_API_KEY",
				"message": "Missing or invalid X-Api-Key header",
			})
			return
		}

		if channel != "ERP" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "X-Routing-Channel must be ERP",
			})
			return
		}

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "invalid request body",
			})
			return
		}

		log.Printf("   subOrderId=%s vendorId=%s message=%q raisedByEmail=%s",
			req.SubOrderID, req.VendorID, req.Message, req.RaisedByEmail)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "vendor complaint raised successfully.",
		})
	}
	http.HandleFunc("/routing/raise-complaint", raiseComplaintHandler)
	http.HandleFunc("/api/v1/routing/raise-complaint", raiseComplaintHandler)
	http.HandleFunc("/fnp-one-support-ticketing/api/v1/routing/raise-complaint", raiseComplaintHandler)

	// 2. Endpoint: tickets routing
	ticketsHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		apiKey := r.Header.Get("X-Api-Key")
		channel := r.Header.Get("X-Routing-Channel")

		log.Printf("📥 Received routing tickets | X-Api-Key=%s | X-Routing-Channel=%s", apiKey, channel)

		if apiKey != "ERP" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "INVALID_API_KEY",
				"message": "Missing or invalid X-Api-Key header",
			})
			return
		}

		if channel != "ERP" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "X-Routing-Channel must be ERP",
			})
			return
		}

		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "invalid request body",
			})
			return
		}

		log.Printf("   subOrderId=%v orderId=%v vendorId=%v incidenceType=%v",
			req["subOrderId"], req["orderId"], req["vendorId"], req["incidenceType"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"message":    "ticket routed successfully",
			"ticketId":   "mock-tkt-123456",
			"subOrderId": req["subOrderId"],
		})
	}
	http.HandleFunc("/api/v1/routing/tickets", ticketsHandler)
	http.HandleFunc("/fnp-one-support-ticketing/api/v1/routing/tickets", ticketsHandler)

	// 3. Endpoint: complaints count
	complaintsHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "invalid request body",
			})
			return
		}

		log.Printf("📥 Received complaints count | vendorId=%v | dateFrom=%v | dateTo=%v",
			req["vendorId"], req["dateFrom"], req["dateTo"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"totalComplaints": 3,
				"byStatus": map[string]interface{}{
					"OPEN":   1,
					"CLOSED": 2,
				},
				"withVendorTicket": 2,
				"vendorTicketByStatus": map[string]interface{}{
					"RESOLVED": 2,
				},
			},
		})
	}

	http.HandleFunc("/erp/tickets/complaints/count", complaintsHandler)
	http.HandleFunc("/fnp-one-support-ticketing/api/v1/erp/tickets/complaints/count", complaintsHandler)
	// 4. Healthcheck endpoints
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Mock SLA server is running!"))
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start HTTP server on PORT environment variable (default 9293)
	port := os.Getenv("PORT")
	if port == "" {
		port = "9293"
	}
	addr := ":" + port
	fmt.Printf("🚀 Mock SLA server running on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
