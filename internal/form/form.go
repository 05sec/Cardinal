// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package form

import (
	"net/http"

	"github.com/flamego/binding"
	"github.com/flamego/flamego"
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/i18n"
)

func Bind(model interface{}) flamego.Handler {
	validate := validator.New()

	return binding.JSON(model, binding.Options{
		ErrorHandler: errorHandler(),
		Validator:    validate,
	})
}

func errorHandler() flamego.Handler {
	return func(c flamego.Context, errors binding.Errors, l *i18n.Locale) {
		c.ResponseWriter().WriteHeader(http.StatusBadRequest)
		c.ResponseWriter().Header().Set("Content-Type", "application/json")

		var errorCode int
		var msg string
		if errors[0].Category == binding.ErrorCategoryDeserialization {
			errorCode = 40000
			msg = l.T("general.error_payload")
		} else {
			errorCode = 40001
			errs := errors[0].Err.(validator.ValidationErrors)
			err := errs[0]

			fieldName := l.T("form." + err.Namespace())

			switch err.Tag() {
			case "required":
				msg = l.T("form.required_error", fieldName)
			case "len":
				msg = l.T("form.len_error", fieldName)
			default:
				msg = err.Error()
			}
		}

		body := map[string]interface{}{
			"error": errorCode,
			"msg":   msg,
		}
		err := jsoniter.NewEncoder(c.ResponseWriter()).Encode(body)
		if err != nil {
			log.Error("Failed to encode response body: %v", err)
		}
	}
}
