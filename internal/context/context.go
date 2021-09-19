// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"strconv"

	"github.com/flamego/flamego"
	jsoniter "github.com/json-iterator/go"
	"github.com/unknwon/com"
	log "unknwon.dev/clog/v2"
)

// Context represents context of a request.
type Context struct {
	flamego.Context
}

func (c *Context) Success(data ...interface{}) error {
	c.ResponseWriter().Header().Set("Content-Type", "application/json")
	c.ResponseWriter().WriteHeader(http.StatusOK)

	var d interface{}
	if len(data) == 1 {
		d = data[0]
	} else {
		d = ""
	}

	err := jsoniter.NewEncoder(c.ResponseWriter()).Encode(
		map[string]interface{}{
			"error": 0,
			"data":  d,
		},
	)
	if err != nil {
		log.Error("Failed to encode: %v", err)
	}
	return nil
}

func (c *Context) ServerError() error {
	return c.Error(http.StatusInternalServerError*100, "Internal server error")
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

// Query queries form parameter.
func (c *Context) Query(name string) string {
	return c.Request().URL.Query().Get(name)
}

// QueryInt returns query result in int type.
func (c *Context) QueryInt(name string) int {
	return com.StrTo(c.Query(name)).MustInt()
}

// QueryInt64 returns query result in int64 type.
func (c *Context) QueryInt64(name string) int64 {
	return com.StrTo(c.Query(name)).MustInt64()
}

// QueryFloat64 returns query result in float64 type.
func (c *Context) QueryFloat64(name string) float64 {
	v, _ := strconv.ParseFloat(c.Query(name), 64)
	return v
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
