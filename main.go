package main

import (
	"log"

	"net/http"

	"encoding/json"

	"io/ioutil"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tealeg/xlsx"
	"gopkg.in/mgo.v2/bson"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func returnAllProducts(w http.ResponseWriter, r *http.Request) {
	db, err := getDB()
	if err != nil {
		json.NewEncoder(w).Encode(err)
		return
	}
	var products []Product
	result := db.Find(&products)
	if result.Error != nil {
		log.Println("Error creating connection pool: ", result.Error)
		json.NewEncoder(w).Encode(result.Error)
		return
	}
	json.NewEncoder(w).Encode(products)
}

func getDB() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=tectoro123 dbname=mydb port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return db, err
}

func returnSingleProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	db, err := getDB()
	if err != nil {
		json.NewEncoder(w).Encode(err)
		return
	}
	var product Product
	result := db.Where("id=?", key).Find(&product)
	if result.Error != nil {
		log.Println("Error creating connection pool: ", result.Error)
		json.NewEncoder(w).Encode(result.Error)
		return
	}
	json.NewEncoder(w).Encode(product)
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var product Product
	json.Unmarshal(reqBody, &product)
	db, err := getDB()
	if err != nil {
		json.NewEncoder(w).Encode(err)
		return
	}
	db.AutoMigrate(&Product{})
	result := db.Create(&product)
	if result.Error != nil {
		log.Println("Error creating connection pool: ", result.Error)
		json.NewEncoder(w).Encode(result.Error)
		return
	}
	json.NewEncoder(w).Encode(product)
}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	vars := mux.Vars(r)
	key := vars["id"]
	var product Product
	json.Unmarshal(reqBody, &product)
	db, err := getDB()
	if err != nil {
		json.NewEncoder(w).Encode(err)
		return
	}
	result := db.Model(&Product{}).Where("id =?", key).Updates(product)
	if result.Error != nil {
		log.Println("Error creating connection pool: ", result.Error)
		json.NewEncoder(w).Encode(result.Error)
		return
	}
	json.NewEncoder(w).Encode(product)
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	db, err := getDB()
	if err != nil {
		json.NewEncoder(w).Encode(err)
		return
	}
	result := db.Where("id = ?", id).Delete(&Product{})
	// result := db.Delete(&Product{}, id)
	if result.Error != nil {
		log.Println("Error creating connection pool: ", result.Error)
		json.NewEncoder(w).Encode(result.Error)
		return
	}
	json.NewEncoder(w).Encode("Record deleted successfully")
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("An error occured while reading the file ", file, " ", err)
	}
	readFile, err := xlsx.OpenBinary(fileContent)
	if err != nil {
		log.Println("An error occured while reading the file content ", err)
	}
	resp, err := readFile.ToSlice()
	headers := bson.M{}
	responseData := make([]map[string]interface{}, 0)
	for idx, row := range resp[0] {
		if idx == 0 {
			for cellIdx, cell := range row {
				headers[strings.TrimSpace(strings.ToLower(cell))] = cellIdx
			}
			continue
		}
		data := prepareData(headers, row)
		responseData = append(responseData, data)
	}
	db, err := getDB()
	db.AutoMigrate(&Product{})
	result := db.Model(&Product{}).Create(responseData)
	if result.Error != nil {
		log.Println("Error creating connection pool: ", result.Error)
		json.NewEncoder(w).Encode(result.Error)
		return
	}
	defer file.Close()
	json.NewEncoder(w).Encode(responseData)
}

func prepareData(fileHeaders map[string]interface{}, row []string) map[string]interface{} {
	data := make(map[string]interface{}, 0)
	for k, v := range fileHeaders {
		if v.(int) < len(row) {
			data[k] = strings.TrimSpace(row[v.(int)])
			continue
		}
		data[k] = ""
	}
	return data
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/all", returnAllProducts)
	myRouter.HandleFunc("/product/{id}", returnSingleProduct).Methods("GET")
	myRouter.HandleFunc("/product", createProduct).Methods("POST")
	myRouter.HandleFunc("/product/{id}", updateProduct).Methods("PUT")
	myRouter.HandleFunc("/product/{id}", deleteProduct).Methods("DELETE")
	myRouter.HandleFunc("/product/upload", uploadFile).Methods("POST")

	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func main() {
	handleRequests()
}
