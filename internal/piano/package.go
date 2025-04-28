// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package piano

import (
	"encoding/json"
	"strings"
)

type AnyResponse struct {
	Code             int               `json:"code"`
	Message          *string           `json:"message"`
	ValidationErrors *ValidationErrors `json:"validation_errors"`
	Raw              json.RawMessage   `json:"-"`
}

type ValidationErrors struct {
	Message string `json:"message"`
}

func (res *AnyResponse) UnmarshalJSON(data []byte) error {
	type response AnyResponse
	var clone response
	err := json.Unmarshal(data, &clone)
	if err != nil {
		return err
	}
	*res = (AnyResponse)(clone)
	res.Raw = data
	return nil
}

func ValidPianoEndpoint(input string) bool {
	predefinedEndpoints := []string{
		"https://sandbox.piano.io/api/v3",
		"https://api-eu.piano.io/api/v3",
		"https://api-au.piano.io/api/v3",
		"https://api-eu.piano.io/api/v3",
		"https://api.piano.io/api/v3",
	}
	exist := false
	for _, url := range predefinedEndpoints {
		if strings.HasPrefix(input, url) {
			exist = true
			break
		}
	}
	return exist
}
