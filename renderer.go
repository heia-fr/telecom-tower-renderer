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
	"encoding/json"
	"fmt"
	"github.com/BlueMasters/tokenauth"
	"github.com/gorilla/mux"
	"github.com/heia-fr/telecom-tower/ledmatrix"
	"github.com/heia-fr/telecom-tower/ledmatrix/font"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
)

type Space struct {
	Len             int    `json:"len"`
	BackgroundColor string `json:"bgColor"`
}

type TextMsg struct {
	Text            string `json:"text"`
	FontSize        int    `json:"fontSize"`
	ForegroundColor string `json:"fgColor"`
	BackgroundColor string `json:"bgColor"`
}

type Matrix struct {
	Rows    int              `json:"rows"`
	Columns int              `json:"columns"`
	Bitmap  ledmatrix.Stripe `json:"bitmap"`
}

var ta = tokenauth.TokenAuth{
	PublicKeyXVar:  "PUBLIC_KEY_X",
	PublicKeyYVar:  "PUBLIC_KEY_Y",
	DatastoreKind:  "ValidTokens",
	MemcachePrefix: "valid-token-",
}

func init() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/renderText", renderText)
	router.HandleFunc("/renderSpace", renderSpace)
	router.HandleFunc("/renderImage", renderImage)
	router.HandleFunc("/join", join)
	http.Handle("/", router)
}

func renderSpace(w http.ResponseWriter, r *http.Request) {
	if err := ta.CheckAuth(r); err != nil {
		http.Error(w, fmt.Sprintf("Access denied: %v", err), 401)
		return
	}

	space := Space{}
	d := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := d.Decode(&space)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), 400)
		return
	}
	var red, green, blue int
	_, err = fmt.Sscanf(space.BackgroundColor, "#%02x%02x%02x", &red, &green, &blue)
	if err != nil {
		http.Error(w, "unable to parse background color", 400)
		return
	}
	bgColor := ledmatrix.RGB(red, green, blue)
	matrix := ledmatrix.NewMatrix(8, 0)
	writer := ledmatrix.NewWriter(matrix)
	writer.Spacer(space.Len, bgColor)
	e := json.NewEncoder(w)
	e.Encode(&matrix)
}

func renderText(w http.ResponseWriter, r *http.Request) {
	if err := ta.CheckAuth(r); err != nil {
		http.Error(w, fmt.Sprintf("Access denied: %v", err), 401)
		return
	}

	msg := TextMsg{}
	d := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := d.Decode(&msg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), 400)
		return
	}

	var red, green, blue int
	_, err = fmt.Sscanf(msg.ForegroundColor, "#%02x%02x%02x", &red, &green, &blue)
	if err != nil {
		http.Error(w, "unable to parse foreground color", 400)
		return
	}
	fgColor := ledmatrix.RGB(red, green, blue)

	_, err = fmt.Sscanf(msg.BackgroundColor, "#%02x%02x%02x", &red, &green, &blue)
	if err != nil {
		http.Error(w, "unable to parse background color", 400)
		return
	}
	bgColor := ledmatrix.RGB(red, green, blue)

	var f font.Font
	if msg.FontSize < 8 {
		f = font.Font6x8
	} else {
		f = font.Font8x8
	}

	matrix := ledmatrix.NewMatrix(8, 0)
	writer := ledmatrix.NewWriter(matrix)
	writer.Spacer(matrix.Columns, 0) // Blank bootstrap
	writer.WriteText(msg.Text, f, fgColor, bgColor)
	e := json.NewEncoder(w)

	e.Encode(&matrix)
}

func renderImage(w http.ResponseWriter, r *http.Request) {
	if err := ta.CheckAuth(r); err != nil {
		http.Error(w, fmt.Sprintf("Access denied: %v", err), 401)
		return
	}

	m, _, err := image.Decode(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid Image: %v", err), 400)
		return
	}
	bounds := m.Bounds()
	if bounds.Max.Y-bounds.Min.Y != 8 {
		http.Error(w, "Invalid Image Size. Must be 8 pixel heigh", 400)
		return
	}

	result := ledmatrix.NewMatrix(bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := m.At(x, y).RGBA()
			result.SetPixel(x-bounds.Min.X, y-bounds.Min.Y, ledmatrix.RGB(int(r>>8), int(g>>8), int(b>>8)))
		}
	}

	e := json.NewEncoder(w)
	e.Encode(&result)
}

func join(w http.ResponseWriter, r *http.Request) {
	if err := ta.CheckAuth(r); err != nil {
		http.Error(w, fmt.Sprintf("Access denied: %v", err), 401)
		return
	}

	var list []Matrix
	d := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := d.Decode(&list)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), 400)
		return
	}
	if len(list) == 0 {
		http.Error(w, "Empty list", 400)
		return
	}
	result := &ledmatrix.Matrix{
		Rows:    list[0].Rows,
		Columns: list[0].Columns,
		Bitmap:  list[0].Bitmap,
	}
	for i := 1; i < len(list); i++ {
		result.Append(&ledmatrix.Matrix{
			Rows:    list[i].Rows,
			Columns: list[i].Columns,
			Bitmap:  list[i].Bitmap,
		})
	}
	e := json.NewEncoder(w)
	e.Encode(&result)
}
