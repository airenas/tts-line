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
		Delete(ID string) (int, error)
	}

	Params struct {
		To       time.Time
		Delete   bool
		Out      io.Writer
		Exporter Exporter
	}
)

type sData struct {
	date time.Time
	d    []*mongodb.TextRecord
}

//Export loads data and saves to writer
// does deletion if param indicates it
func Export(p Params) error {
	goapp.Log.Infof("Exporting data")
	goapp.Log.Infof("Loading data")
	data, err := p.Exporter.All()
	if err != nil {
		return errors.Wrap(err, "can't load data")
	}
	goapp.Log.Infof("Filtering data")
	data = filterData(data, p.To)
	goapp.Log.Infof("Sorting data")
	data = sortData(data)
	goapp.Log.Infof("Writing data. %d items", len(data))

	if err = writeData(data, p.Out); err != nil {
		return errors.Wrap(err, "can't write data")
	}
	if p.Delete {
		goapp.Log.Infof("Deleting data")
		if err = deleteData(data, p.Exporter.Delete); err != nil {
			return errors.Wrap(err, "can't delete records")
		}
	}
	return nil
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

func filterData(data []*mongodb.TextRecord, to time.Time) []*mongodb.TextRecord {
	if to.IsZero() {
		goapp.Log.Info("No filter")
		return data
	}
	goapp.Log.Infof("Filter records newer than '%s'", to.Format("2006-01-02"))
	tm := make(map[string]bool)
	for _, d := range data {
		if utils.RequestTypeEnum(d.Type) == utils.RequestOriginal && to.After(d.Created) {
			tm[d.ID] = true
		}
	}
	res := make([]*mongodb.TextRecord, 0)
	for _, d := range data {
		if to.After(d.Created) || tm[d.ID] {
			res = append(res, d)
		}
	}
	return res
}

func deleteData(data []*mongodb.TextRecord, deleteFunc func(string) (int, error)) error {
	tm := make(map[string]bool)
	c := 0
	for _, d := range data {
		if !tm[d.ID] {
			cd, err := deleteFunc(d.ID)
			if err != nil {
				return errors.Wrapf(err, "can't delete ID %s", d.ID)
			}
			c += cd
			tm[d.ID] = true
		}
	}
	goapp.Log.Infof("Deleted %d records", c)
	return nil
}

func writeData(data []*mongodb.TextRecord, wr io.Writer) error {
	je := json.NewEncoder(wr)
	je.SetEscapeHTML(false)
	_, err := wr.Write([]byte{'['})
	if err != nil {
		return errors.Wrap(err, "Can't write")
	}
	comma := false
	for _, d := range data {
		if comma {
			_, err = wr.Write([]byte{','})
			if err != nil {
				return errors.Wrap(err, "Can't write")
			}
		}
		err = je.Encode(d)
		if err != nil {
			return errors.Wrap(err, "Can't marshal")
		}
		comma = true
	}
	_, err = wr.Write([]byte{']'})
	if err != nil {
		return errors.Wrap(err, "Can't write")
	}
	return nil
}
