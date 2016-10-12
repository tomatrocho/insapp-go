package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/smtp"
	"time"
)

// User defines how to model a User
type User struct {
	ID          bson.ObjectId   `bson:"_id,omitempty"`
	Name        string          `json:"name"`
	Username    string          `json:"username"`
	Description string          `json:"description"`
	Email       string          `json:"email"`
	EmailPublic bool            `json:"emailpublic"`
	Promotion   string          `json:"promotion"`
	Gender 			string					`json:"gender"`
	Events      []bson.ObjectId `json:"events"`
	PostsLiked  []bson.ObjectId `json:"postsliked"`
}

// Users is an array of User
type Users []User

// AddUser will add the given user from JSON body to the database
func AddUser(user User) User {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("insapp").C("user")
	db.Insert(user)
	var result User
	db.Find(bson.M{"username": user.Username}).One(&result)
	return result
}

// UpdateUser will update the user link to the given ID,
// with the field of the given user, in the database
func UpdateUser(id bson.ObjectId, user User) User {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("insapp").C("user")
	userID := bson.M{"_id": id}
	change := bson.M{"$set": bson.M{
		"name":        user.Name,
		"description": user.Description,
		"email": 			 user.Email,
		"emailpublic": user.EmailPublic,
		"promotion":   user.Promotion,
		"gender"	:		 user.Gender,
	}}
	db.Update(userID, change)
	var result User
	db.Find(bson.M{"_id": id}).One(&result)
	return result
}

// DeleteUser will delete the given user from the database
func DeleteUser(user User) User {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("insapp").C("user")
	DeleteCredentalsForUser(user.ID)
	DeleteNotificationsForUser(user.ID)
	DeleteNotificationTokenForUser(user.ID)
	for _, eventId := range user.Events{
		RemoveParticipant(eventId, user.ID)
	}
	for _, postId := range user.PostsLiked{
		DislikePostWithUser(postId, user.ID)
	}
	DeleteTagsForUser(user.ID)
	DeleteCommentsForUser(user.ID)
	db.RemoveId(user.ID)
	var result User
	db.FindId(user.ID).One(result)
	return result
}

// GetUser will return an User object from the given ID
func GetUser(id bson.ObjectId) User {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("insapp").C("user")
	var result User
	db.FindId(id).One(&result)
	return result
}

// LikePost will add the postID to the list of liked post
// of the user linked to the given id
func LikePost(id bson.ObjectId, postID bson.ObjectId) User {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("insapp").C("user")
	userID := bson.M{"_id": id}
	change := bson.M{"$addToSet": bson.M{
		"postsliked": postID,
	}}
	db.Update(userID, change)
	var result User
	db.Find(bson.M{"_id": id}).One(&result)
	return result
}

// DislikePost will remove the postID from the list of liked
// post of the user linked to the given id
func DislikePost(id bson.ObjectId, postID bson.ObjectId) User {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("insapp").C("user")
	userID := bson.M{"_id": id}
	change := bson.M{"$pull": bson.M{
		"postsliked": postID,
	}}
	db.Update(userID, change)
	var result User
	db.Find(bson.M{"_id": id}).One(&result)
	return result
}

// AddEventToUser will add the eventID to the list
// of the user's event linked to the given id
func AddEventToUser(id bson.ObjectId, eventID bson.ObjectId) User {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("insapp").C("user")
	userID := bson.M{"_id": id}
	change := bson.M{"$addToSet": bson.M{
		"events": eventID,
	}}
	db.Update(userID, change)
	var result User
	db.Find(bson.M{"_id": id}).One(&result)
	return result
}

// RemoveEventFromUser will remove the eventID from the list
// of the user's event linked to the given id
func RemoveEventFromUser(id bson.ObjectId, eventID bson.ObjectId) User {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("insapp").C("user")
	userID := bson.M{"_id": id}
	change := bson.M{"$pull": bson.M{
		"events": eventID,
	}}
	db.Update(userID, change)
	var result User
	db.Find(bson.M{"_id": id}).One(&result)
	return result
}

func SearchUser(username string) Users {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("insapp").C("user")
	var result Users
	db.Find(bson.M{"$or" : []interface{}{
		bson.M{"username" : bson.M{ "$regex" : bson.RegEx{`^.*` + username + `.*`, "i"}}}, bson.M{"name" : bson.M{ "$regex" : bson.RegEx{`^.*` + username + `.*`, "i"}}}}}).All(&result)
	return result
}


func ReportUser(id bson.ObjectId) {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("insapp").C("user")
	var user User
	db.Find(bson.M{"_id": id}).One(&user)
	SendEmail("aeir@insa-rennes.fr", "Un utilisateur a été reporté sur Insapp",
		"Cet utilisateur a été reporté le " + time.Now().String() +
		"\n\n" + user.ID.Hex() + "\n" + user.Username + "\n" + user.Name + "\n" + user.Description)
}


func SendEmail(to string, subject string, body string) {
  from := "insapp.contact@gmail.com"
	pass := "PASSWORD"
	cc := "insapp.contact@gmail.com"
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
    "Cc: " + cc + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))
}
