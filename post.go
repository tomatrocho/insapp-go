package insapp

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Post defines how to model a Post
type Post struct {
	ID             bson.ObjectId   `bson:"_id,omitempty"`
	Title          string          `json:"title"`
	Association    bson.ObjectId   `json:"association"`
	Description    string          `json:"description"`
	Date           time.Time       `json:"date"`
	Likes          []bson.ObjectId `json:"likes"`
	Comments       Comments        `json:"comments"`
	Promotions     []string        `json:"promotions"`
	Plateforms     []string        `json:"plateforms"`
	Image          string          `json:"image"`
	ImageSize      bson.M          `json:"imageSize"`
	NoNotification bool            `json:"nonotification"`
}

// Posts is an array of Post
type Posts []Post

// AddPost will add the given Post to the database
func AddPost(post Post) Post {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("post")

	_ = db.Insert(post)
	var result Post
	_ = db.Find(bson.M{"title": post.Title, "date": post.Date}).One(&result)
	AddPostToAssociation(result.Association, result.ID)

	return result
}

// UpdatePost will update the post linked to the given ID,
// with the field of the given post, in the database
func UpdatePost(id bson.ObjectId, post Post) Post {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("post")

	postID := bson.M{"_id": id}
	change := bson.M{"$set": bson.M{
		"title":          post.Title,
		"description":    post.Description,
		"image":          post.Image,
		"plateforms":     post.Plateforms,
		"promotions":     post.Promotions,
		"imageSize":      post.ImageSize,
		"nonotification": post.NoNotification,
	}}
	_ = db.Update(postID, change)

	var result Post
	_ = db.Find(bson.M{"_id": id}).One(&result)

	return result
}

// DeletePost will delete the given Post from the database
func DeletePost(post Post) Post {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("post")

	_ = db.RemoveId(post.ID)
	var result Post
	_ = db.FindId(post.ID).One(result)
	DeleteNotificationsForPost(post.ID)
	RemovePostFromAssociation(post.Association, post.ID)
	for _, userID := range post.Likes {
		DislikePost(userID, post.ID)
	}

	return result
}

// GetPost will return a Post object from the given ID
func GetPost(id bson.ObjectId) Post {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("post")

	var result Post
	_ = db.FindId(id).One(&result)

	return result
}

// GetPosts will return an array of Posts
func GetPosts() Posts {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("post")

	var result Posts
	_ = db.Find(bson.M{}).Sort("-date").All(&result)

	return result
}

// GetLatestPosts will return an array of the last N Posts
func GetLatestPosts(number int) Posts {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("post")

	var result Posts
	_ = db.Find(bson.M{}).Sort("-date").Limit(number).All(&result)

	return result
}

// GetPostsForAssociation returns an array of Posts from the given association ID
func GetPostsForAssociation(id bson.ObjectId) Posts {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("post")

	var result Posts
	_ = db.Find(bson.M{"association": id}).Sort("-date").All(&result)

	return result
}

func SearchPost(name string) Posts {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("post")

	var result Posts
	_ = db.Find(bson.M{"$or": []interface{}{
		bson.M{"title": bson.M{"$regex": bson.RegEx{Pattern: `^.*` + name + `.*`, Options: "i"}}}, bson.M{"description": bson.M{"$regex": bson.RegEx{Pattern: `^.*` + name + `.*`, Options: "i"}}}}}).All(&result)

	return result
}

// LikePostWithUser will add the user to the list of
// user that liked the post (cf. Likes field)
func LikePostWithUser(id bson.ObjectId, userID bson.ObjectId) (Post, User) {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("post")

	postID := bson.M{"_id": id}
	change := bson.M{"$addToSet": bson.M{
		"likes": userID,
	}}
	_ = db.Update(postID, change)

	var post Post
	_ = db.Find(bson.M{"_id": id}).One(&post)
	user := LikePost(userID, post.ID)

	return post, user
}

// DislikePostWithUser will remove the user to the list of
// users that liked the post (cf. Likes field)
func DislikePostWithUser(id bson.ObjectId, userID bson.ObjectId) (Post, User) {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("post")

	postID := bson.M{"_id": id}
	change := bson.M{"$pull": bson.M{
		"likes": userID,
	}}
	_ = db.Update(postID, change)

	var post Post
	_ = db.Find(bson.M{"_id": id}).One(&post)
	user := DislikePost(userID, post.ID)

	return post, user
}
