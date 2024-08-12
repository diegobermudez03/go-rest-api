package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

var db *sql.DB

func initDB() {
    var err error
    connStr := os.Getenv("DATABASE_URL")
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
}

func getMachineTypes(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, name, description FROM machine_type")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var machineTypes []map[string]interface{}
    for rows.Next() {
        var id int
        var name, description string
        rows.Scan(&id, &name, &description)
        machineTypes = append(machineTypes, map[string]interface{}{
            "id":          id,
            "name":        name,
            "description": description,
        })
    }

    json.NewEncoder(w).Encode(machineTypes)
}

func getMachines(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, name, serial_number, machine_type_id, status FROM machine")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var machines []map[string]interface{}
    for rows.Next() {
        var id int
        var name, serialNumber string
        var machineTypeID int
        var status string
        rows.Scan(&id, &name, &serialNumber, &machineTypeID, &status)
        machines = append(machines, map[string]interface{}{
            "id":                    id,
            "name":                  name,
            "serial_number":         serialNumber,
            "machine_type_id":       machineTypeID,
            "status":                status,
        })
    }

    json.NewEncoder(w).Encode(machines)
}

func createMachineType(w http.ResponseWriter, r *http.Request) {
    var machineType struct {
        Name        string `json:"name"`
        Description string `json:"description"`
    }
    if err := json.NewDecoder(r.Body).Decode(&machineType); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    _, err := db.Exec("INSERT INTO machine_type (name, description) VALUES ($1, $2)", machineType.Name, machineType.Description)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}

func createMachine(w http.ResponseWriter, r *http.Request) {
    var machine struct {
        Name                 string `json:"name"`
        SerialNumber         string `json:"serial_number"`
        MachineTypeID        int    `json:"machine_type_id"`
        Status               string `json:"status"`
    }
    if err := json.NewDecoder(r.Body).Decode(&machine); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    _, err := db.Exec("INSERT INTO machine (name, serial_number, machine_type_id, status) VALUES ($1, $2, $3, $4)",
        machine.Name, machine.SerialNumber, machine.MachineTypeID, machine.Status)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
    initDB()
    defer db.Close()

    r := mux.NewRouter()

    // Machine Type Endpoints
    r.HandleFunc("/machine-types", getMachineTypes).Methods("GET")
    r.HandleFunc("/machine-types", createMachineType).Methods("POST")

    // Machine Endpoints
    r.HandleFunc("/machines", getMachines).Methods("GET")
    r.HandleFunc("/machines", createMachine).Methods("POST")

    http.Handle("/", r)
    log.Println("Server running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
