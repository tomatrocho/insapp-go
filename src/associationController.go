package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"image"
	_ "image/jpeg"
	_ "image/png"

	"gopkg.in/mgo.v2/bson"

	"github.com/gorilla/mux"
)

func GetMyAssociationController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assocationID := vars["id"]
	var res = GetMyAssociations(bson.ObjectIdHex(assocationID))
	json.NewEncoder(w).Encode(res)
}

// GetAssociationController will answer a JSON of the association
// linked to the given id in the URL
func GetAssociationController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assocationID := vars["id"]
	var res = GetAssociation(bson.ObjectIdHex(assocationID))
	json.NewEncoder(w).Encode(res)
}

// GetAllAssociationsController will answer a JSON of all associations
func GetAllAssociationsController(w http.ResponseWriter, r *http.Request) {
	var res = GetAllAssociation()
	json.NewEncoder(w).Encode(res)
}

func CreateUserForAssociationController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assocationID := vars["id"]
	var res = GetAssociation(bson.ObjectIdHex(assocationID))

	decoder := json.NewDecoder(r.Body)
	var user AssociationUser
	decoder.Decode(&user)

	user.Association = res.ID
	user.Username = res.Email
	user.Password = GetMD5Hash(user.Password)
	AddAssociationUser(user)
	json.NewEncoder(w).Encode(res)
}

// AddAssociationController will answer a JSON of the
// brand new created association (from the JSON Body)
func AddAssociationController(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var association Association
	decoder.Decode(&association)
	res := AddAssociation(association)
	json.NewEncoder(w).Encode(res)
}

// UpdateAssociationController will answer the JSON of the
// modified association (from the JSON Body)
func UpdateAssociationController(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var association Association
	decoder.Decode(&association)
	vars := mux.Vars(r)
	assocationID := vars["id"]
	res := UpdateAssociation(bson.ObjectIdHex(assocationID), association)
	json.NewEncoder(w).Encode(res)
}

// DeleteAssociationController will answer a JSON of an
// empty association if the deletation has succeed
func DeleteAssociationController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	res := DeleteAssociation(bson.ObjectIdHex(vars["id"]))
	json.NewEncoder(w).Encode(res)
}

// AddProfileImageAssociationController will set the profile image of the association and return the association
func AddProfileImageAssociationController(w http.ResponseWriter, r *http.Request) {
	fileName := UploadImage(r)
	if fileName == "error" {
		w.Header().Set("status", "400")
		fmt.Fprintln(w, "{}")
	} else {
		vars := mux.Vars(r)
		res := SetImageAssociation(bson.ObjectIdHex(vars["id"]), fileName, true)
		json.NewEncoder(w).Encode(res)
	}
}

// AddCoverImageAssociationController will set the cover image of the association and return the association
func AddCoverImageAssociationController(w http.ResponseWriter, r *http.Request) {
	fileName := UploadImage(r)
	if fileName == "error" {
		w.Header().Set("status", "400")
		fmt.Fprintln(w, "{}")
	} else {
		vars := mux.Vars(r)
		res := SetImageAssociation(bson.ObjectIdHex(vars["id"]), fileName, false)
		json.NewEncoder(w).Encode(res)
	}
}

// UploadImage will manage the upload image from a POST request
func UploadImage(r *http.Request) string {
	fmt.Println("upload image called")
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return "error"
	}
	defer file.Close()

	fileName := RandomString(40)
	f, err := os.OpenFile("./img/"+fileName+".png", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "error"
	}

	defer f.Close()
	io.Copy(f, file)

	return fileName
}

func GetImageDimension(fileName string) (int, int) {
    file, err1 := os.Open("./img/"+fileName+".png")
    if err1 != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err1)
    }
    image, _, err2 := image.DecodeConfig(file)
    if err2 != nil {
        fmt.Fprintf(os.Stderr, "%s: %v\n", fileName, err2)
    }
    return image.Width, image.Height
}

// RandomString generates a randomString (y)
func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
