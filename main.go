package main

import (
    "encoding/json"
    "log"
    "net/http"
    "math/rand"
    "strconv"
    "github.com/gorilla/mux"
)

type Book struct {
    ID    string  `json:"id"`
    Isbn  string  `json:"isbn"`
    Title string  `json:"title"`
    Author *Author `json:"author"`
}

type Author struct {
    Firstname string `json:"firstname"`
    Lastname  string `json:"lastname"`
}

var books []Book

func getBooks(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)

    for _, item := range books {
        if item.ID == params["id"] {
            json.NewEncoder(w).Encode(item)
            return
        }
    }

    json.NewEncoder(w).Encode(&Book{})
}

func createBook(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var book Book
    _ = json.NewDecoder(r.Body).Decode(&book)
    book.ID = strconv.Itoa(rand.Intn(1000000)) // unsafe id
    books = append(books, book)
    json.NewEncoder(w).Encode(book)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)

    for index, item := range books {
        if item.ID == params["id"] {
            books = append(books[:index], books[index+1:]...)
            var book Book
            _ = json.NewDecoder(r.Body).Decode(&book)
            book.ID = params["id"]
            books = append(books, book)
            json.NewEncoder(w).Encode(book)
            return
        }
    }

    json.NewEncoder(w).Encode(books)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)

    for index, item := range books {
        if item.ID == params["id"] {
            books = append(books[:index], books[index+1:]...)
            break
        }
    }

    json.NewEncoder(w).Encode(books)
}

func main() {
    router := mux.NewRouter()

    books = append(books, Book{ID: "1", Isbn: "446841", Title: "Book", Author: &Author{Firstname: "John", Lastname: "Doe"}})

    router.HandleFunc("/api/books", getBooks).Methods("GET")
    router.HandleFunc("/api/books/{id}", getBook).Methods("GET")
    router.HandleFunc("/api/books", createBook).Methods("POST")
    router.HandleFunc("/api/books/{id}", updateBook).Methods("PUT")
    router.HandleFunc("/api/books/{id}", deleteBook).Methods("DELETE")

    log.Fatal(http.ListenAndServe(":3000", router))
}