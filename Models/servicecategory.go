package Models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServiceCategory struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
}
