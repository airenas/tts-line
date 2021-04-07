package mongodb

import (
	"context"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TextSaver saves text to mongo DB
type TextSaver struct {
	SessionProvider *SessionProvider
}

//NewTextSaver creates TextSaver instance
func NewTextSaver(sessionProvider *SessionProvider) (*TextSaver, error) {
	if sessionProvider == nil {
		return nil, errors.New("No session provider")
	}
	f := TextSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// Save text to DB
func (ss *TextSaver) Save(req, text string, reqType utils.RequestTypeEnum, tags []string) error {
	goapp.Log.Infof("Saving ID %s", req)

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(textTable).Collection(textTable)
	res := toRecord(req, text, reqType, tags)
	_, err = c.InsertOne(ctx, res)
	return err
}

// All loads all records
func (ss *TextSaver) All() ([]*TextRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute) // increase to retrieve all records
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(textTable).Collection(textTable)
	cursor, err := c.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "Can't get data")
	}
	defer cursor.Close(ctx)
	res := make([]*TextRecord, 0)
	for cursor.Next(ctx) {
		var key TextRecord
		if err = cursor.Decode(&key); err != nil {
			return nil, errors.Wrap(err, "Can't get key")
		}
		res = append(res, &key)
	}
	return res, nil
}

// LoadText by ID and type
func (ss *TextSaver) LoadText(requestID string, reqType utils.RequestTypeEnum) (string, error) {
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return "", err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(textTable).Collection(textTable)
	var res TextRecord
	err = c.FindOne(ctx, bson.M{
		"$and": []bson.M{
			{"id": sanitize(requestID)},
			{"type": int(reqType)},
		}}).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", utils.ErrNoRecord
		}
		return "", errors.Wrap(err, "Can't get data")
	}
	return res.Text, nil
}

func toRecord(req, text string, reqType utils.RequestTypeEnum, tags []string) *TextRecord {
	res := &TextRecord{}
	res.ID = req
	res.Text = text
	res.Type = int(reqType)
	res.Created = time.Now()
	res.Tags = tags
	return res
}
