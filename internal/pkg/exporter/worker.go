package exporter

import (
	"encoding/json"
	"io"
	"sort"
	"time"

	"github.com/airenas/tts-line/internal/pkg/mongodb"
	"github.com/airenas/tts-line/internal/pkg/utils"

	"github.com/airenas/go-app/pkg/goapp"

	"github.com/pkg/errors"
)

type (
	//Exporter retrieves all data from DB
	Exporter interface {
		All() ([]*mongodb.TextRecord, error)
	}
)

type sData struct {
	date time.Time
	d    []*mongodb.TextRecord
}

//Export lodas data and saves to witer
func Export(exp Exporter, wr io.Writer) error {
	goapp.Log.Infof("Exporting data")
	goapp.Log.Infof("Loading data")
	data, err := exp.All()
	if err != nil {
		return errors.Wrap(err, "Can't load data")
	}
	goapp.Log.Infof("Sorting data")
	data = sortData(data)
	goapp.Log.Infof("Writing data. %d items", len(data))
	je := json.NewEncoder(wr)
	return je.Encode(data)
}

func sortData(data []*mongodb.TextRecord) []*mongodb.TextRecord {
	tm := make(map[string]*sData)
	for _, d := range data {
		fd := tm[d.ID]
		if fd == nil {
			fd = &sData{date: d.Created, d: []*mongodb.TextRecord{d}}
			tm[d.ID] = fd
		} else {
			if utils.RequestTypeEnum(d.Type) == utils.RequestOriginal {
				fd.date = d.Created
			}
			fd.d = append(fd.d, d)
		}
	}
	tl := make([]*sData, 0)
	for _, d := range tm {
		sort.Slice(d.d, func(i, j int) bool { return d.d[i].Type < d.d[j].Type })
		tl = append(tl, d)
	}
	sort.Slice(tl, func(i, j int) bool { return tl[i].date.Before(tl[j].date) })
	res := make([]*mongodb.TextRecord, 0)
	for _, d := range tl {
		res = append(res, d.d...)
	}
	return res
}
