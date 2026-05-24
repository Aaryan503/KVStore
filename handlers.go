package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func createTable(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var req CreateTableRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	table, err := db.createTable(name, req.Columns)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(table)
}

func insertRow(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "tableName")
	rowID := chi.URLParam(r, "rowId")

	var row Row

	err := json.NewDecoder(r.Body).Decode(&row)
	if err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	insertedRow, err := db.insertRow(tableName, rowID, row)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(insertedRow)
}

func deleteTable(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "name")
	err := db.deleteTable(tableName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("deleted table"))
}

func deleteRow(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "tableName")
	rowId := chi.URLParam(r, "rowId")
	err := db.deleteRow(tableName, rowId)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("deleted row"))
}

func getTables(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var result []*Table
	result = db.getTables()
	json.NewEncoder(w).Encode(result)
}

func getTable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tableName := chi.URLParam(r, "tableName")
	table, err := db.getTable(tableName)
	if err != nil {
		http.Error(w, err.Error(), 404)
	}
	json.NewEncoder(w).Encode(table)
}

func getRow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tableName := chi.URLParam(r, "tableName")
	rowId := chi.URLParam(r, "rowId")
	row, err := db.getRow(tableName, rowId)
	if err != nil {
		http.Error(w, err.Error(), 404)
	}
	json.NewEncoder(w).Encode(row)
}
