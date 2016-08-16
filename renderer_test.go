package renderer

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRenderText(t *testing.T) {

	txt := TextMsg{
		Text:            "+",
		ForegroundColor: "#000100",
		BackgroundColor: "#010000",
		FontSize:        6,
	}

	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.Encode(&txt)

	req, err := http.NewRequest("POST", "/renderText", &b)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	renderText(w, req)

	assert.Equal(t, 200, w.Code)

	dec := json.NewDecoder(w.Body)
	assert.NoError(t, err)

	var result Matrix
	err = dec.Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, 8, result.Rows)
	assert.Equal(t, 6, result.Columns)
	assert.Equal(t, 8*6, len(result.Bitmap))

	assert.EqualValues(t, 1<<16, result.Bitmap[0])
	assert.EqualValues(t, 1<<8, result.Bitmap[11])
}
