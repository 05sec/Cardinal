// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package i18n

import (
	"github.com/flamego/flamego"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
	"golang.org/x/text/language"

	"github.com/vidar-team/Cardinal/internal/context"
)

type Locale struct {
	Tag language.Tag
	*i18n.I18n
}

func (l *Locale) T(key string, args ...interface{}) string {
	return string(l.I18n.T(l.Tag.String(), key, args...))
}

func I18n() flamego.Handler {
	// TODO go embed
	yamlBackend := yaml.New("./locales")
	translations := yamlBackend.LoadTranslations()
	i18n := i18n.New(yamlBackend)

	tags := make([]language.Tag, 0)
	for _, tr := range translations {
		tags = append(tags, language.Raw.Make(tr.Locale))
	}
	matcher := language.NewMatcher(tags)

	return func(ctx context.Context) {
		acceptLanguages := ctx.Request().Header.Get("Accept-Language")
		tags, _, _ := language.ParseAcceptLanguage(acceptLanguages)
		tag, _, _ := matcher.Match(tags...)

		locale := &Locale{
			Tag:  tag,
			I18n: i18n,
		}
		ctx.Map(locale)
	}
}
