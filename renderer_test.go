// Copyright 2016 Jacques Supcik, HEIA-FR
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package renderer

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/go-uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"net/http/httptest"
	"os"
	"testing"
)

type token struct {
	Present bool
}

var validUUID, invalidUUID string
var validToken, invalidToken string

func insertToken(ctx context.Context, uuid string) error {
	t := token{
		Present: true,
	}
	key := datastore.NewKey(ctx, "ValidTokens", uuid, 0, nil)
	_, err := datastore.Put(ctx, key, &t)
	return err
}

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

	inst, err := aetest.NewInstance(nil)
	assert.NoError(t, err)

	req, err := inst.NewRequest("POST", "/renderText", &b)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

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

func TestRenderImage(t *testing.T) {
	b, err := os.Open("logo.png")
	assert.NoError(t, err)

	inst, err := aetest.NewInstance(nil)
	assert.NoError(t, err)

	req, err := inst.NewRequest("POST", "/renderImage", b)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w := httptest.NewRecorder()
	renderImage(w, req)

	assert.Equal(t, 200, w.Code)

	dec := json.NewDecoder(w.Body)
	assert.NoError(t, err)

	var result Matrix
	err = dec.Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, 8, result.Rows)
	assert.Equal(t, 8, result.Columns)
	assert.Equal(t, 8*8, len(result.Bitmap))

	assert.EqualValues(t, 0, result.Bitmap[0])
	assert.EqualValues(t, 2596560, result.Bitmap[11])
}

func TestRenderSpace(t *testing.T) {
	txt := Space{
		Len:             13,
		BackgroundColor: "#010000",
	}

	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.Encode(&txt)

	inst, err := aetest.NewInstance(nil)
	assert.NoError(t, err)

	req, err := inst.NewRequest("POST", "/renderSpace", &b)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w := httptest.NewRecorder()
	renderSpace(w, req)

	assert.Equal(t, 200, w.Code)

	dec := json.NewDecoder(w.Body)
	assert.NoError(t, err)

	var result Matrix
	err = dec.Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, 8, result.Rows)
	assert.Equal(t, 13, result.Columns)
	assert.Equal(t, 8*13, len(result.Bitmap))

	assert.EqualValues(t, 1<<16, result.Bitmap[0])
	assert.EqualValues(t, 1<<16, result.Bitmap[11])
}

func TestJoin(t *testing.T) {
	txt1 := TextMsg{
		Text:            "+",
		ForegroundColor: "#000001",
		BackgroundColor: "#010000",
		FontSize:        8,
	}

	txt2 := TextMsg{
		Text:            "+",
		ForegroundColor: "#000010",
		BackgroundColor: "#100000",
		FontSize:        6,
	}

	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.Encode(&txt1)

	inst, err := aetest.NewInstance(nil)
	assert.NoError(t, err)

	req, err := inst.NewRequest("POST", "/renderText", &b)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w := httptest.NewRecorder()
	renderText(w, req)

	assert.Equal(t, 200, w.Code)

	dec := json.NewDecoder(w.Body)
	assert.NoError(t, err)

	var result1 Matrix
	err = dec.Decode(&result1)
	assert.NoError(t, err)

	assert.Equal(t, 8, result1.Rows)
	assert.Equal(t, 8, result1.Columns)
	assert.Equal(t, 8*8, len(result1.Bitmap))

	assert.EqualValues(t, 1<<16, result1.Bitmap[0])
	assert.EqualValues(t, 1<<0, result1.Bitmap[11])

	b.Reset()
	enc.Encode(&txt2)

	req, err = inst.NewRequest("POST", "/renderText", &b)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w = httptest.NewRecorder()
	renderText(w, req)

	assert.Equal(t, 200, w.Code)

	dec = json.NewDecoder(w.Body)
	assert.NoError(t, err)

	var result2 Matrix
	err = dec.Decode(&result2)
	assert.NoError(t, err)

	assert.Equal(t, 8, result2.Rows)
	assert.Equal(t, 6, result2.Columns)
	assert.Equal(t, 8*6, len(result2.Bitmap))

	assert.EqualValues(t, 1<<20, result2.Bitmap[0])
	assert.EqualValues(t, 1<<4, result2.Bitmap[11])

	list := []Matrix{result1, result2}

	b.Reset()
	enc.Encode(&list)

	req, err = inst.NewRequest("POST", "/join", &b)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w = httptest.NewRecorder()
	join(w, req)

	assert.Equal(t, 200, w.Code)

	dec = json.NewDecoder(w.Body)
	assert.NoError(t, err)

	var result Matrix
	err = dec.Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, 8, result.Rows)
	assert.Equal(t, (8 + 6), result.Columns)
	assert.Equal(t, 8*(8+6), len(result.Bitmap))

	assert.EqualValues(t, 1<<16, result.Bitmap[0])
	assert.EqualValues(t, 1<<0, result.Bitmap[11])

	assert.EqualValues(t, 1<<20, result.Bitmap[8*8+0])
	assert.EqualValues(t, 1<<4, result.Bitmap[8*8+11])
}

func TestRenderSpaceErrors(t *testing.T) {
	b := bytes.NewBufferString("GARBAGE")

	inst, err := aetest.NewInstance(nil)
	assert.NoError(t, err)

	req, err := inst.NewRequest("POST", "/renderSpace", b)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w := httptest.NewRecorder()
	renderSpace(w, req)

	assert.Equal(t, 400, w.Code)

	txt := Space{
		Len:             12,
		BackgroundColor: "BADCOLOR",
	}

	var b1 bytes.Buffer
	enc := json.NewEncoder(&b1)
	enc.Encode(&txt)

	req, err = inst.NewRequest("POST", "/renderSpace", &b1)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w = httptest.NewRecorder()
	renderSpace(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestRenderTextErrors(t *testing.T) {
	b := bytes.NewBufferString("GARBAGE")

	inst, err := aetest.NewInstance(nil)
	assert.NoError(t, err)

	req, err := inst.NewRequest("POST", "/renderText", b)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w := httptest.NewRecorder()
	renderText(w, req)

	assert.Equal(t, 400, w.Code)

	txt := TextMsg{
		Text:            "+",
		ForegroundColor: "BADCOLOR",
		BackgroundColor: "#010000",
		FontSize:        6,
	}

	var b1 bytes.Buffer
	enc := json.NewEncoder(&b1)
	enc.Encode(&txt)

	req, err = inst.NewRequest("POST", "/renderText", &b1)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w = httptest.NewRecorder()
	renderText(w, req)

	txt = TextMsg{
		Text:            "+",
		ForegroundColor: "#ff00CC",
		BackgroundColor: "BADCOLOR",
		FontSize:        6,
	}

	b1.Reset()
	enc.Encode(&txt)

	req, err = inst.NewRequest("POST", "/renderText", &b1)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w = httptest.NewRecorder()
	renderText(w, req)
}

func TestRenderImageErrors(t *testing.T) {
	b, err := os.Open("LICENSE")
	assert.NoError(t, err)

	inst, err := aetest.NewInstance(nil)
	assert.NoError(t, err)

	req, err := inst.NewRequest("POST", "/renderImage", b)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w := httptest.NewRecorder()
	renderImage(w, req)

	assert.Equal(t, 400, w.Code)

	b, err = os.Open("badsize.png")
	assert.NoError(t, err)

	req, err = inst.NewRequest("POST", "/renderImage", b)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w = httptest.NewRecorder()
	renderImage(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestJoinError(t *testing.T) {

	b := bytes.NewBufferString("GARBAGE")

	inst, err := aetest.NewInstance(nil)
	assert.NoError(t, err)

	req, err := inst.NewRequest("POST", "/join", b)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w := httptest.NewRecorder()
	join(w, req)

	assert.Equal(t, 400, w.Code)

	list := []Matrix{}

	var b1 bytes.Buffer
	enc := json.NewEncoder(&b1)

	enc.Encode(&list)

	req, err = inst.NewRequest("POST", "/join", &b1)
	assert.NoError(t, err)

	insertToken(appengine.NewContext(req), validUUID)
	req.Header.Add("Authorization", "Bearer "+validToken)

	w = httptest.NewRecorder()
	join(w, req)

	assert.Equal(t, 400, w.Code)
}

func init() {
	var err error
	validUUID, err = uuid.GenerateUUID()
	if err != nil {
		panic(err)
	}

	invalidUUID, err = uuid.GenerateUUID()
	if err != nil {
		panic(err)
	}

	goodPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	os.Setenv("PUBLIC_KEY_X", fmt.Sprintf("%X", goodPrivateKey.X))
	os.Setenv("PUBLIC_KEY_Y", fmt.Sprintf("%X", goodPrivateKey.Y))

	validToken, err = jwt.NewWithClaims(jwt.SigningMethodES256, &jwt.StandardClaims{
		Audience: "basic",
		Subject:  "Jacques",
		Issuer:   "BlueMasters",
		Id:       validUUID,
	}).SignedString(goodPrivateKey)

	if err != nil {
		panic(err)
	}

	invalidToken, err = jwt.NewWithClaims(jwt.SigningMethodES256, &jwt.StandardClaims{
		Audience: "basic",
		Subject:  "Jacques",
		Issuer:   "BlueMasters",
		Id:       invalidUUID,
	}).SignedString(goodPrivateKey)

	if err != nil {
		panic(err)
	}

}
