package Models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name"`
	Price           float64            `bson:"price" json:"price"`
	Description     string             `bson:"description" json:"description"`
	ServiceCategory primitive.ObjectID `bson:"servicecategory" json:"servicecategory"`
	ImageURL        string             `bson:"imageurl" json:"imageurl"`
}
