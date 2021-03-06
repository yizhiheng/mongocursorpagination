package bson

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/require"
)

type item struct {
	ID        bson.ObjectId `json:"id" bson:"_id"`
	Name      string        `json:"name" bson:"name"`
	UserID    string        `json:"userId" bson:"userId,omitempty"`
	CreatedAt time.Time     `json:"createdAt" bson:"createdAt"`
}

func TestGenerateCursorQuery(t *testing.T) {
	var cases = []struct {
		name                    string
		shouldSecondarySortOnID bool
		paginatedField          string
		comparisonOp            string
		cursorFieldValues       []interface{}
		expectedQuery           map[string]interface{}
		expectedErr             error
	}{
		{
			"error when wrong number of cursor field values specified and shouldSecondarySortOnID is true",
			true,
			"name",
			"$gt",
			[]interface{}{"abc"},
			nil,
			errors.New("wrong number of cursor field values specified"),
		},
		{
			"error when wrong number of cursor field values specified and shouldSecondarySortOnID is false",
			false,
			"_id",
			"$lt",
			[]interface{}{},
			nil,
			errors.New("wrong number of cursor field values specified"),
		},
		{
			"return appropriate cursor query when shouldSecondarySortOnID is true",
			true,
			"name",
			"$gt",
			[]interface{}{"test item", "123"},
			map[string]interface{}{"$or": []map[string]interface{}{
				{"name": map[string]interface{}{"$gt": "test item"}},
				{"$and": []map[string]interface{}{
					{"name": map[string]interface{}{"$eq": "test item"}},
					{"_id": map[string]interface{}{"$gt": "123"}}},
				},
			}},
			nil,
		},
		{
			"return appropriate cursor query when shouldSecondarySortOnID is false",
			false,
			"_id",
			"$lt",
			[]interface{}{"123"},
			map[string]interface{}{"_id": map[string]interface{}{"$lt": "123"}},
			nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := GenerateCursorQuery(tc.shouldSecondarySortOnID, tc.paginatedField, tc.comparisonOp, tc.cursorFieldValues)
			require.Equal(t, tc.expectedQuery, query)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestFindStructFieldNameByBsonTag(t *testing.T) {
	var cases = []struct {
		name                    string
		structType              reflect.Type
		tag                     string
		expectedStructFieldName string
	}{
		{
			"return struct field name when matching bson tag specified",
			reflect.TypeOf(item{}),
			"name",
			"Name",
		},
		{
			"return struct field name when tag has additional flags",
			reflect.TypeOf(item{}),
			"userId",
			"UserID",
		},
		{
			"return empty struct field name when a non matching bson tag specified",
			reflect.TypeOf(item{}),
			"notastructfield",
			"",
		},
		{
			"return empty struct field name when tag is empty",
			reflect.TypeOf(item{}),
			"",
			"",
		},
		{
			"return empty struct field name when structType is nil",
			nil,
			"name",
			"",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			structFieldName := FindStructFieldNameByBsonTag(tc.structType, tc.tag)
			require.Equal(t, tc.expectedStructFieldName, structFieldName)
		})
	}
}
