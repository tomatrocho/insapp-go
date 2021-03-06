package insapp

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// NotificationUser defines how to model a NotificationUser
type NotificationUser struct {
	ID     bson.ObjectId `bson:"_id,omitempty"`
	UserId bson.ObjectId `json:"userid"`
	Token  string        `json:"token" bson:",omitempty"`
	Os     string        `json:"os"`
}

// Notification defines how to model a Notification
type Notification struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Sender   bson.ObjectId `json:"sender"`
	Receiver bson.ObjectId `json:"receiver"`
	Content  bson.ObjectId `json:"content"`
	Comment  Comment       `json:"comment,omitempty" bson:",omitempty"`
	Message  string        `json:"message"`
	Seen     bool          `json:"seen"`
	Date     time.Time     `json:"date"`
	Type     string        `json:"type"`
}

type Notifications []Notification

func GetNotificationUserForUser(userID bson.ObjectId) NotificationUser {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("notification_user")

	var result NotificationUser
	db.Find(bson.M{"userid": userID}).One(&result)

	return result
}

func CreateOrUpdateNotificationUser(user NotificationUser) {
	if len(user.Token) == 0 {
		return
	}

	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("notification_user")

	res, _ := db.Find(bson.M{"token": user.Token}).Count()
	if res > 0 {
		db.Update(bson.M{"token": user.Token}, bson.M{"$set": bson.M{"token": nil}})
	}
	res, _ = db.Find(bson.M{"userid": user.UserId}).Count()
	if res > 0 {
		db.Update(bson.M{"userid": user.UserId}, bson.M{"$set": bson.M{"token": user.Token, "os": user.Os}})
	} else {
		db.Insert(user)
	}
}

func AddNotification(notification Notification) Notification {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("notification")

	notification.ID = bson.NewObjectId()
	notification.Date = time.Now()
	notification.Seen = false
	db.Insert(notification)
	return notification
}

func GetNotificationsForUser(userID bson.ObjectId) Notifications {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("notification")

	var result Notifications
	db.Find(bson.M{"receiver": userID}).Sort("-date").Limit(30).All(&result)

	return result
}

func GetUnreadNotificationsForUser(userID bson.ObjectId) Notifications {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("notification")

	var result Notifications
	db.Find(bson.M{"receiver": userID, "seen": false}).Sort("-date").Limit(30).All(&result)
	return result
}

func ReadNotificationForUser(userID bson.ObjectId, notifID bson.ObjectId) Notifications {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("notification")

	db.Update(bson.M{"_id": notifID}, bson.M{"$set": bson.M{"seen": true}})

	return GetNotificationsForUser(userID)
}

func DeleteNotificationsForUser(id bson.ObjectId) {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("notification")

	db.RemoveAll(bson.M{"receiver": id})
}

func DeleteNotificationsForComment(id bson.ObjectId) {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("notification")

	db.RemoveAll(bson.M{"comment._id": id})
}

func DeleteNotificationsForPost(id bson.ObjectId) {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("notification")

	db.RemoveAll(bson.M{"content": id})
}

func DeleteNotificationsForEvent(id bson.ObjectId) {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("notification")

	db.RemoveAll(bson.M{"content": id})
}

func DeleteNotificationTokenForUser(id bson.ObjectId) {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("notification_user")

	db.RemoveAll(bson.M{"userid": id})
}
