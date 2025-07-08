package stemplate

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

var testTime time.Time

func init() {
	const longForm = "2006/01/02 15:04"
	var err error
	testTime, err = time.Parse(longForm, "2016/07/03 12:34")
	if err != nil {
		panic(err)
	}
}

func TestDowncase(t *testing.T) {
	assert.Equal(t, "hello", downcase("HeLlO"))
}

func TestFormatTimeWithMinute(t *testing.T) {
	assert.Equal(t, "July 3, 2016 12:34", formatTimeWithMinute(testTime))
}

func TestFormatTimeYearMonth(t *testing.T) {
	assert.Equal(t, "July 2016", formatTimeYearMonth(testTime))
}

func TestInKM(t *testing.T) {
	assert.Equal(t, 2.342, inKM(2342.0)) //nolint:testifylint
}

func TestMod(t *testing.T) {
	assert.Equal(t, 0, mod(2, 2))
	assert.Equal(t, 1, mod(3, 2))
}

func TestMonthName(t *testing.T) {
	assert.Equal(t, "July", monthName(time.July))
}

func TestNumberWithDelimiter(t *testing.T) {
	assert.Equal(t, "123", numberWithDelimiter(',', 123))
	assert.Equal(t, "1,234", numberWithDelimiter(',', 1234))
	assert.Equal(t, "12,345", numberWithDelimiter(',', 12345))
	assert.Equal(t, "123,456", numberWithDelimiter(',', 123456))
	assert.Equal(t, "1,234,567", numberWithDelimiter(',', 1234567))
}

func TestPace(t *testing.T) {
	d := 60 * time.Second

	// Easiest case: 1000 m ran in 60 seconds which is 1:00 per km.
	assert.Equal(t, "1:00", pace(1000.0, d))

	// Fast: 2000 m ran in 60 seconds which is 0:30 per km.
	assert.Equal(t, "0:30", pace(2000.0, d))

	// Slow: 133 m ran in 60 seconds which is 7:31 per km.
	assert.Equal(t, "7:31", pace(133.0, d))
}

func TestRound(t *testing.T) {
	assert.Equal(t, 0.0, round(0.2)) //nolint:testifylint
	assert.Equal(t, 1.0, round(0.8)) //nolint:testifylint
	assert.Equal(t, 1.0, round(0.5)) //nolint:testifylint
}

func TestToStars(t *testing.T) {
	assert.Empty(t, toStars(0))
	assert.Equal(t, "★ ", toStars(1))
	assert.Equal(t, "★ ★ ★ ★ ★ ", toStars(5))
}

func TestURLBaseExt(t *testing.T) {
	assert.Empty(t, urlBaseExt("https://example.com/video"))
	assert.Equal(t, "jpg", urlBaseExt("https://example.com/image.JPG"))
	assert.Equal(t, "mp4", urlBaseExt("https://example.com/video.mp4"))
	assert.Equal(t, "webm", urlBaseExt("https://example.com/video.webm"))
}

func TestURLBaseFile(t *testing.T) {
	assert.Equal(t, "video", urlBaseFile("https://example.com/video"))
	assert.Equal(t, "image.JPG", urlBaseFile("https://example.com/image.JPG"))
	assert.Equal(t, "video.mp4", urlBaseFile("https://example.com/video.mp4"))
	assert.Equal(t, "video.webm", urlBaseFile("https://example.com/video.webm"))
}
