// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package context

import (
	"net/http"

	"github.com/flamego/flamego"
	jsoniter "github.com/json-iterator/go"
	log "unknwon.dev/clog/v2"
)

// Context represents context of a request.
type Context struct {
	flamego.Context
}

func (c *Context) Success(data interface{}) error {
	c.ResponseWriter().Header().Set("Content-Type", "application/json")
	c.ResponseWriter().WriteHeader(http.StatusOK)

	err := jsoniter.NewEncoder(c.ResponseWriter()).Encode(
		map[string]interface{}{
			"error": 0,
			"msg":   "success",
			"data":  data,
		},
	)
	if err != nil {
		log.Error("Failed to encode: %v", err)
	}
	return nil
}

func (c *Context) ServerError() error {
	return c.Error(http.StatusInternalServerError, "Internal server error")
}

func (c *Context) Error(errorCode uint, message string) error {
	statusCode := int(errorCode / 100)

	c.ResponseWriter().Header().Set("Content-Type", "application/json")
	c.ResponseWriter().WriteHeader(statusCode)

	err := jsoniter.NewEncoder(c.ResponseWriter()).Encode(
		map[string]interface{}{
			"error": errorCode,
			"msg":   message,
		},
	)
	if err != nil {
		log.Error("Failed to encode: %v", err)
	}
	return nil
}

// Contexter initializes a classic context for a request.
func Contexter() flamego.Handler {
	return func(ctx flamego.Context) {
		c := Context{
			Context: ctx,
		}

		c.Map(c)
	}
}
