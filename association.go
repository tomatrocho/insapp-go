package insapp

import (
	"gopkg.in/mgo.v2/bson"
)

// Association defines the model of a Association
type Association struct {
	ID              bson.ObjectId   `bson:"_id,omitempty"`
	Name            string          `json:"name"`
	Email           string          `json:"email"`
	Description     string          `json:"description"`
	Events          []bson.ObjectId `json:"events"`
	Posts           []bson.ObjectId `json:"posts"`
	Palette         [][]int         `json:"palette"`
	SelectedColor   int             `json:"selectedcolor"`
	Profile         string          `json:"profile"`
	ProfileUploaded string          `json:"profileuploaded"`
	Cover           string          `json:"cover"`
	BgColor         string          `json:"bgcolor"`
	FgColor         string          `json:"fgcolor"`
}

// Associations is an array of Association
type Associations []Association

// AddAssociationUser will add the given AssociationUser to the database
func AddAssociationUser(user AssociationUser) {
	session := GetMongoSession()
	defer session.Close()

	db := session.DB("insapp").C("association_user")
	db.Insert(user)
}

// AddAssociation will add the given Association to the database
func AddAssociation(association Association) Association {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association")

	db.Insert(association)
	var result Association
	db.Find(bson.M{"name": association.Name}).One(&result)

	return result
}

// UpdateAssociation will update the given Association link to the given ID,
// with the field of the given Association, in the database
func UpdateAssociation(id bson.ObjectId, association Association) Association {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association")

	if association.ProfileUploaded != "" {
		association.Profile, _ = ResizeImage(association.ProfileUploaded, 256, 256)
	}

	associationID := bson.M{"_id": id}
	change := bson.M{"$set": bson.M{
		"name":            association.Name,
		"email":           association.Email,
		"description":     association.Description,
		"profile":         association.Profile,
		"profileuploaded": association.ProfileUploaded,
		"cover":           association.Cover,
		"palette":         association.Palette,
		"selectedcolor":   association.SelectedColor,
		"bgcolor":         association.BgColor,
		"fgcolor":         association.FgColor,
	}}

	db.Update(associationID, change)
	var result Association
	db.Find(bson.M{"_id": id}).One(&result)

	return result
}

// DeleteAssociation will delete the given association from the database
func DeleteAssociation(id bson.ObjectId) Association {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association")

	association := GetAssociation(id)
	for _, eventID := range association.Events {
		DeleteEvent(GetEvent(eventID))
	}
	for _, postID := range association.Posts {
		DeletePost(GetPost(postID))
	}

	db.RemoveId(id)
	var result Association
	db.FindId(id).One(result)

	return result
}

// GetAssociation will return an Association object from the given ID
func GetAssociation(id bson.ObjectId) Association {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association")

	var result Association
	db.FindId(id).One(&result)

	return result
}

// GetAssociationFromEmail will return an Association object from the given email
func GetAssociationFromEmail(email string) Association {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association")

	var result Association
	db.Find(bson.M{"email": email}).One(&result)

	return result
}

// GetAllAssociations will return an array of all the existing Association, hidding "Menu" association and sort by name asc
func GetAllAssociations() Associations {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association")

	var result Associations
	db.Find(bson.M{"name": bson.M{"$ne": "Menu"}}).Sort("name").All(&result)

	return result
}

// GetMyAssociations will return an array of all ID from owned existing Association
func GetMyAssociations(id bson.ObjectId) []bson.ObjectId {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association_user")

	var result []AssociationUser
	db.Find(bson.M{"owner": id}).All(&result)
	var res []bson.ObjectId
	for _, association := range result {
		res = append(res, association.Association)
	}

	return res
}

// SearchAssociation return an array of all Association found with the given search string.
func SearchAssociation(name string) Associations {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association")

	var result Associations
	db.Find(bson.M{"$or": []interface{}{
		bson.M{"name": bson.M{"$regex": bson.RegEx{`^.*` + name + `.*`, "i"}}},
		bson.M{"description": bson.M{"$regex": bson.RegEx{`^.*` + name + `.*`, "i"}}}}}).All(&result)

	return result
}

// AddEventToAssociation will add the given event ID to the given association
func AddEventToAssociation(id bson.ObjectId, event bson.ObjectId) Association {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association")

	associationID := bson.M{"_id": id}
	change := bson.M{"$addToSet": bson.M{
		"events": event,
	}}

	db.Update(associationID, change)
	var result Association
	db.Find(bson.M{"_id": id}).One(&result)

	return result
}

// RemoveEventFromAssociation will remove the given event ID from the given association
func RemoveEventFromAssociation(id bson.ObjectId, event bson.ObjectId) Association {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association")

	associationID := bson.M{"_id": id}
	change := bson.M{"$pull": bson.M{
		"events": event,
	}}

	db.Update(associationID, change)
	var result Association
	db.Find(bson.M{"_id": id}).One(&result)

	return result
}

func AddPostToAssociation(id bson.ObjectId, post bson.ObjectId) Association {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association")

	associationID := bson.M{"_id": id}
	change := bson.M{"$addToSet": bson.M{
		"posts": post,
	}}

	db.Update(associationID, change)
	var result Association
	db.Find(bson.M{"_id": id}).One(&result)

	return result
}

func RemovePostFromAssociation(id bson.ObjectId, post bson.ObjectId) Association {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association")

	associationID := bson.M{"_id": id}
	change := bson.M{"$pull": bson.M{
		"posts": post,
	}}

	db.Update(associationID, change)
	var result Association
	db.Find(bson.M{"_id": id}).One(&result)

	return result
}

// GetAssociationUser return the AssociationUser object with the given ID.
func GetAssociationUser(id bson.ObjectId) AssociationUser {
	session := GetMongoSession()
	defer session.Close()
	db := session.DB("insapp").C("association_user")

	var result AssociationUser
	db.Find(bson.M{"association": id}).One(&result)

	return result
}
