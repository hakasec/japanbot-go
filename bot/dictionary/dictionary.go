package dictionary

import (
	"io"

	jmdict "github.com/hakasec/jmdict-go"
)

type Dictionary struct {
	Index     map[string][]*jmdict.Entry
	IndexByID map[string][]*jmdict.Entry

	*jmdict.JMdict
}

func (d *Dictionary) createIndex() {
	for i, entry := range d.Entries {
		for _, kanji := range entry.KanjiElements {
			if _, ok := d.Index[kanji.Phrase]; !ok {
				d.Index[kanji.Phrase] = make([]*jmdict.Entry, 0)
			}
			d.Index[kanji.Phrase] = append(d.Index[kanji.Phrase], &d.Entries[i])
		}
		for _, reading := range entry.ReadingElements {
			if _, ok := d.Index[reading.Phrase]; !ok {
				d.Index[reading.Phrase] = make([]*jmdict.Entry, 0)
			}
			if _, ok := d.Index[reading.PhraseNoKanji]; !ok {
				d.Index[reading.PhraseNoKanji] = make([]*jmdict.Entry, 0)
			}
			d.Index[reading.Phrase] = append(d.Index[reading.Phrase], &d.Entries[i])
			d.Index[reading.PhraseNoKanji] = append(d.Index[reading.PhraseNoKanji], &d.Entries[i])
		}
		if _, ok := d.IndexByID[entry.EntryID]; !ok {
			d.IndexByID[entry.EntryID] = append(d.IndexByID[entry.EntryID], &d.Entries[i])
		}
	}
}

func Load(r io.Reader) (*Dictionary, error) {
	var err error
	d := &Dictionary{}
	d.JMdict, err = jmdict.Load(r)
	if err != nil {
		return nil, err
	}

	d.Index = make(map[string][]*jmdict.Entry)
	d.IndexByID = make(map[string][]*jmdict.Entry)
	d.createIndex()

	return d, nil
}
