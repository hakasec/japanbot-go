package dictionary

import (
	"io"

	jmdict "github.com/hakasec/jmdict-go"
)

type Dictionary struct {
	Index     map[string]*jmdict.Entry
	IndexByID map[string]*jmdict.Entry

	*jmdict.JMdict
}

func (d *Dictionary) createIndex() {
	for i, entry := range d.Entries {
		for _, kanji := range entry.KanjiElements {
			d.Index[kanji.Phrase] = &d.Entries[i]
		}
		for _, reading := range entry.ReadingElements {
			d.Index[reading.Phrase] = &d.Entries[i]
			d.Index[reading.PhraseNoKanji] = &d.Entries[i]
		}
		d.IndexByID[entry.EntryID] = &d.Entries[i]
	}
}

func Load(r io.Reader) (*Dictionary, error) {
	var err error
	d := &Dictionary{}
	d.JMdict, err = jmdict.Load(r)
	if err != nil {
		return nil, err
	}

	d.Index = make(map[string]*jmdict.Entry)
	d.IndexByID = make(map[string]*jmdict.Entry)
	d.createIndex()

	return d, nil
}
