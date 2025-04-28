// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package syntax

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"terraform-provider-piano/internal/piano"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func AnyResponseFrom(response *http.Response, diagnostics *diag.Diagnostics) (*piano.AnyResponse, error) {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		diagnostics.AddError("IO Error", fmt.Sprintf("Unable to read licensee, got error: %s", err))
		return nil, err
	}
	anyResponse := piano.AnyResponse{}
	err = json.Unmarshal(body, &anyResponse)
	if err != nil {
		diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage, got error: %s", err))
		return nil, err
	}
	if anyResponse.Code != 0 {
		diagnostics.AddError(fmt.Sprintf("Status Error: %d: %s", anyResponse.Code, *anyResponse.Message), string(anyResponse.Raw))
		return nil, errors.New("Status Error")
	}
	return &anyResponse, err
}
