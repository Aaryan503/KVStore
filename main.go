package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

type Item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Store struct {
	mu      sync.RWMutex
	items   map[string]Item
	walFile *os.File
}

type WAL struct {
	Operation string
	Key       string
	Value     string
	Timestamp time.Time
}

var store Store

func (s *Store) getAll() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Item
	for _, item := range s.items {
		result = append(result, item)
	}
	return result
}

func (s *Store) getItem(key string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, exists := s.items[key]
	if !exists {
		return Item{}, false
	}
	return item, true
}

func (s *Store) putItem(item *Item, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	op := WAL{
		Operation: "PUT",
		Key:       key,
		Value:     item.Value,
		Timestamp: time.Now(),
	}

	err := s.appendWAL(op)
	if err != nil {
		return err
	}
	item.Key = key
	s.items[key] = *item
	return nil
}

func (s *Store) deleteItem(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.items[key]
	if !exists {
		return false
	}
	op := WAL{
		Operation: "DELETE",
		Key:       key,
		Timestamp: time.Now(),
	}
	err := s.appendWAL(op)
	if err != nil {
		return false
	}
	delete(s.items, key)
	return true
}

func (s *Store) appendWAL(op WAL) error {
	jsonBytes, err := json.Marshal(op)
	if err != nil {
		return err
	}
	_, err = s.walFile.Write(append(jsonBytes, '\n'))
	if err != nil {
		return err
	}
	err = s.walFile.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) loadWAL() error {
	_, err := s.walFile.Seek(0, 0)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(s.walFile)

	for {
		var op WAL
		err := decoder.Decode(&op)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch op.Operation {
		case "PUT":
			s.items[op.Key] = Item{
				Key:   op.Key,
				Value: op.Value,
			}
		case "DELETE":
			delete(s.items, op.Key)
		}
	}
	return nil
}

func main() {
	wal, err := os.OpenFile(
		"wal.log",
		os.O_CREATE|os.O_APPEND|os.O_RDWR,
		0644,
	)
	if err != nil {
		panic(err)
	}
	store = Store{
		items:   make(map[string]Item),
		walFile: wal,
	}
	err = store.loadWAL()
	if err != nil {
		panic(err)
	}
	r := chi.NewRouter()
	r.Get("/items", getAllItems)
	r.Get("/items/{key}", getItem)
	r.Put("/items/{key}", putItem)
	r.Delete("/items/{key}", deleteItem)
	fmt.Println("Server is running on port 8080")
	defer wal.Close()
	err = http.ListenAndServe(":8080", r)
}

func getAllItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(store.getAll())
}

func getItem(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	item, exists := store.getItem(key)
	if !exists {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func putItem(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	newerr := store.putItem(&item, key)
	if newerr != nil {
		http.Error(w, "storage failure", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	exists := store.deleteItem(key)
	if !exists {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}
	w.Write([]byte("deleted"))
}
