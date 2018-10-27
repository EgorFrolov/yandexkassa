package yandexkassa

import "fmt"

const (
	ErrorInvalidRequest      = "invalid_request"
	ErrorNotSupported        = "not_supported"
	ErrorInvalidCredentials  = "invalid_credentials"
	ErrorForbidden           = "forbidden"
	ErrorNotFound            = "not_found"
	ErrorTooManyRequests     = "too_many_requests"
	ErrorInternalServerError = "internal_server_error"
)

//Структуры для использования в PaymentRequest, Payment, RefundRequest, Refund
type Amount struct {
	Value    string `json:"value"`    //Сумма в выбранной валюте. Выражается в виде строки и пишется через точку, например 10.00. Количество знаков после точки зависит от выбранной валюты.
	Currency string `json:"currency"` //Код валюты в формате ISO-4217. Должен соответствовать валюте вашего аккаунта
}

type Receipt struct {
	Items         []Item `json:"items"`           //Список товаров в заказе
	TaxSystemCode int64  `json:"tax_system_code"` //Система налогообложения магазина
	Phone         string `json:"phone"`           //Телефон пользователя для отправки чека. Указывается в формате ITU-T E.164, например 79000000000
	Email         string `json:"email"`           //Электронная почта пользователя для отправки чека
}

type Item struct {
	Description string `json:"description"` //Название товара
	Quantity    string `json:"quantity"`    //Количество
	Amount      Amount `json:"amount"`      //Цена товара
	VatCode     int    `json:"vat_code"`    //Ставка НДС. Возможные значения — числа от 1 до 6
}

type Kassa struct {
	ShopID         int64
	SecretKey      string
	IdempotenceKey string
}

type Error struct {
	Type        string `json:"type"`        //тип ошибки (e.g. error)
	ID          string `json:"id"`          //ID ошибки (e.g.) ab5a11cd-13cc-4e33-af8b-75a74e18dd09
	Code        string `json:"code"`        //Название ошибки (код) (e.g.) invalid_request
	Description string `json:"description"` //описание самой ошибки (e.g.) Idempotence key duplicated
	Parameter   string `json:"parameter"`   //указывает на параметр, из-за которого возникла ошибка (e.g.) Idempotence-Key
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %s param: %s desc: %s (id: %s)", e.Code, e.Parameter, e.Description, e.ID)
}

func IsYandexError(err error) (*Error, bool) {
	if err, ok := err.(*Error); ok {
		return err, true
	} else {
		return err, false
	}
}
