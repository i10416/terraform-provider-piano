// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package piano

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
		"https://api-ap.piano.io/api/v3",
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

func AnyResponseFrom(response *http.Response) (*AnyResponse, error) {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	anyResponse := AnyResponse{}
	err = json.Unmarshal(body, &anyResponse)
	if err != nil {
		return nil, err
	}
	return &anyResponse, err
}

func SuccessfulResponseFrom(response *http.Response, onError func(summary string, detail string)) (*AnyResponse, error) {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		onError("IO Error", fmt.Sprintf("Unable to read body, got error: %e", err))
		return nil, err
	}
	anyResponse := AnyResponse{}
	err = json.Unmarshal(body, &anyResponse)
	if err != nil {
		onError("Decode Error", fmt.Sprintf("Unable to decode body as AnyResponse, got error: %e", err))
		return nil, err
	}
	if anyResponse.Code != 0 {
		onError(fmt.Sprintf("Status Error: %d: %s", anyResponse.Code, *anyResponse.Message), string(anyResponse.Raw))
		return nil, errors.New("status error")
	}
	return &anyResponse, err
}
