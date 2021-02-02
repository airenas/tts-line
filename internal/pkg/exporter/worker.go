package exporter

import (
	"encoding/json"
	"io"
	"sort"

	"github.com/airenas/tts-line/internal/pkg/mongodb"

	"github.com/airenas/go-app/pkg/goapp"

	"github.com/pkg/errors"
)

type (
	//Exporter retrieves all data from DB
	Exporter interface {
		All() ([]*mongodb.TextRecord, error)
	}
)

//Export lodas data and saves to witer
func Export(exp Exporter, wr io.Writer) error {
	goapp.Log.Infof("Exporting data")
	goapp.Log.Infof("Loading data")
	data, err := exp.All()
	if err != nil {
		return errors.Wrap(err, "Can't load data")
	}
	goapp.Log.Infof("Sorting data")
	sort.Slice(data, func(i, j int) bool { return compare(data[i], data[j]) })
	goapp.Log.Infof("Writing data")
	je := json.NewEncoder(wr)
	return je.Encode(data)
}

//NewRouter creates the router for HTTP service
func compare(d1, d2 *mongodb.TextRecord) bool {
	if d1.ID < d2.ID {
		return true
	} else if d1.ID > d2.ID {
		return false
	}
	if d1.Type < d2.Type {
		return true
	} else if d1.Type > d2.Type {
		return false
	}
	return d1.Created.Before(d2.Created)
}
