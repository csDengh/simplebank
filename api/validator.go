package api

import (
	"github.com/csdengh/cur_blank/utils"
	"github.com/go-playground/validator/v10"
)

var currencyValid validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return utils.IsSupportedCurrency(currency)
	}
	return false
}
