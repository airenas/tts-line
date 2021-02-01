package mongodb

import (
	"context"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

// TextSaver saves text to mongo DB
type TextSaver struct {
	SessionProvider *SessionProvider
}

//NewTextSaver creates TextSaver instance
func NewTextSaver(sessionProvider *SessionProvider) (*TextSaver, error) {
	f := TextSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// Save text to DB
func (ss *TextSaver) Save(req, text string, reqType utils.RequestTypeEnum) error {
	goapp.Log.Infof("Saving ID %s", req)

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(textTable).Collection(textTable)
	res := toRecord(req, text, reqType)
	_, err = c.InsertOne(ctx, res)
	return err
}

func toRecord(req, text string, reqType utils.RequestTypeEnum) *textRecord {
	res := &textRecord{}
	res.ID = req
	res.Text = text
	res.Type = int(reqType)
	res.Created = time.Now()
	return res
}
