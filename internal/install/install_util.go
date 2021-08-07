// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package install

import (
	"strconv"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

var DateTimeLayout = "2006-01-02 15:04:05"

func inputDateTime(label string, validateFunc ...func(time.Time) error) (time.Time, error) {
	validate := func(time.Time) error {
		return nil
	}
	if len(validateFunc) != 0 {
		validate = validateFunc[0]
	}

	prompt := promptui.Prompt{
		Label: label,
		Validate: func(s string) error {
			t, err := time.Parse(DateTimeLayout, s)
			if err != nil {
				return errors.Wrap(err, "parse time")
			}
			return validate(t)
		},
		Default: time.Now().Format(DateTimeLayout),
	}

	dateResult, err := prompt.Run()
	if err != nil {
		return time.Time{}, errors.Wrap(err, "run prompt")
	}

	t, err := time.Parse(DateTimeLayout, dateResult)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "parse time")
	}
	return t, nil
}

func inputInt(label string, defaultValue int, validateFunc ...func(int) error) (int, error) {
	validate := func(int) error {
		return nil
	}
	if len(validateFunc) != 0 {
		validate = validateFunc[0]
	}

	prompt := promptui.Prompt{
		Label: label,
		Validate: func(s string) error {
			val, err := strconv.Atoi(s)
			if err != nil {
				return errors.Wrap(err, "convert")
			}
			return validate(val)
		},
		Default: strconv.Itoa(defaultValue),
	}

	intResult, err := prompt.Run()
	if err != nil {
		return 0, errors.Wrap(err, "run prompt")
	}

	val, err := strconv.Atoi(intResult)
	if err != nil {
		return 0, errors.Wrap(err, "convert")
	}

	return val, nil
}

func inputConfirm(label string, defaultVal ...bool) (bool, error) {
	var defaultValue bool
	if len(defaultVal) != 0 {
		defaultValue = defaultVal[0]
	}

	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
		Default: func(val bool) string {
			if val {
				return "y"
			}
			return "N"
		}(defaultValue),
	}

	_, err := prompt.Run()
	if err != nil {
		// FYI: https://github.com/manifoldco/promptui/issues/81
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, errors.Wrap(err, "run prompt")
	}
	return true, nil
}

func inputString(label, defaultValue string, validateFunc ...func(string) error) (string, error) {
	validate := func(string) error {
		return nil
	}
	if len(validateFunc) != 0 {
		validate = validateFunc[0]
	}

	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
		Default:  defaultValue,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", errors.Wrap(err, "run prompt")
	}
	return result, nil
}

func inputPassword(label string, validateFunc ...func(string) error) (string, error) {
	validate := func(string) error {
		return nil
	}
	if len(validateFunc) != 0 {
		validate = validateFunc[0]
	}

	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
		Mask:     '*',
	}

	result, err := prompt.Run()
	if err != nil {
		return "", errors.Wrap(err, "run prompt")
	}
	return result, nil
}
