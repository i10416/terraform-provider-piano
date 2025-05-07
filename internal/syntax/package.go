// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package syntax

import (
	"net/http"
	"terraform-provider-piano/internal/piano"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func SuccessfulResponseFrom(response *http.Response, diagnostics *diag.Diagnostics) (*piano.AnyResponse, error) {
	return piano.SuccessfulResponseFrom(response, func(summary, detail string) {
		diagnostics.AddError(summary, detail)
	})
}
