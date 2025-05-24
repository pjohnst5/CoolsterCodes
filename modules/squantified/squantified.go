package squantified

import (
	"encoding"
	"html/template"
	"sort"
	"time"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mmarkdown"
	"github.com/brandur/modulir/modules/mtoml"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private types
//
//
//
//////////////////////////////////////////////////////////////////////////////

//
// Goodreads
//

// Reading is a single Goodreads book stored to a TOML file.
type Reading struct {
	Authors       []*ReadingAuthor `toml:"authors"`
	ID            int              `toml:"id"`
	ISBN          string           `toml:"isbn"`
	ISBN13        string           `toml:"isbn13"`
	NumPages      int              `toml:"num_pages"`
	PublishedYear int              `toml:"published_year"`
	ReadAt        time.Time        `toml:"read_at"`
	Rating        int              `toml:"rating"`
	Review        string           `toml:"review"`
	ReviewHTML    template.HTML    `toml:"-"`
	ReviewID      int              `toml:"review_id"`
	Title         string           `toml:"title"`

	// AuthorsDisplay is just the names of all authors combined together for
	// display on a page.
	AuthorsDisplay string `toml:"-"`
}

// ReadingAuthor is a single Goodreads author stored to a TOML file.
type ReadingAuthor struct {
	ID   int    `toml:"id"`
	Name string `toml:"name"`
}

// Verifies interface compliance.
var _ encoding.TextUnmarshaler = &ReadingAuthor{}

// Only kicks in if the author value is a single string. Otherwise the full
// object is unmarshaled into the struct.
func (a *ReadingAuthor) UnmarshalText(data []byte) error {
	a.Name = string(data)
	return nil
}

// ReadingDB is a database of Goodreads readings stored to a TOML file.
type ReadingDB struct {
	Readings []*Reading `toml:"readings"`
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private functions
//
//
//
//////////////////////////////////////////////////////////////////////////////

func combineAuthors(authors []*ReadingAuthor) string {
	if len(authors) == 0 {
		return ""
	}

	if len(authors) == 1 {
		return authors[0].Name
	}

	display := ""

	for i, author := range authors {
		if i == len(authors)-1 {
			display += " & "
		} else if i > 0 {
			display += ", "
		}

		display += author.Name
	}

	return display
}

func GetReadingsData(c *modulir.Context, target string) ([]*Reading, error) {
	var readingDB ReadingDB

	err := retryOnce(c, func() error {
		return mtoml.ParseFile(c, target, &readingDB)
	})
	if err != nil {
		return nil, err
	}

	// Sort in reverse chronological order. Books should be roughly sorted
	// like this already, but they're sorted by review ID, which may be out
	// of order compared to the read date.
	sort.Slice(readingDB.Readings, func(i, j int) bool {
		return readingDB.Readings[i].ReadAt.After(readingDB.Readings[j].ReadAt)
	})

	for _, reading := range readingDB.Readings {
		reading.AuthorsDisplay = combineAuthors(reading.Authors)

		// Empty reviews written before 2020. These are poorly written (more
		// than usual even) and often contained spoilers since I used them like
		// notes.
		if reading.ReadAt.Year() < 2020 {
			reading.Review = ""
		} else {
			reading.ReviewHTML = template.HTML(string(mmarkdown.Render(c, []byte(reading.Review))))
		}
	}

	return readingDB.Readings, nil
}

// Data files (especially Twitter's) can be quite large, and if we having
// something like Vim writing to one, our file watcher may notice the change
// before Vim is finished its write. This causes ioutil to read only a
// partially written file, and the TOML unmarshal below it to subsequently
// fail.
//
// Do some hacky protection against this by retrying once when we encounter an
// error. The process of trying to decode TOML the first time should take
// easily enough time to let Vim finish writing, so we'll pick up the full file
// on the second pass.
//
// Note that this only ever a problem on incremental rebuilds and will never be
// needed otherwise.
func retryOnce(c *modulir.Context, f func() error) error {
	var err error
	for range 2 {
		err = f()
		if err != nil {
			c.Log.Errorf("Errored, but retrying once: %v", err)
			continue
		}
		break
	}
	return err
}
