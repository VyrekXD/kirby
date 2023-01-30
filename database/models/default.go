package models

import (
	"github.com/kamva/mgm/v3"
)

type DefaultModel struct {
	ID             string `json:"id" bson:"_id,omitempty"`
	mgm.DateFields `bson:",inline"`
}

func (f *DefaultModel) PrepareID(id interface{}) (interface{}, error) {
	return id, nil
}

func (f *DefaultModel) GetID() interface{} {
	return f.ID
}

func (f *DefaultModel) SetID(id interface{}) {
	f.ID = id.(string)
}
