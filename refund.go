package yandexkassa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	RefundStatusCanceled  = "canceled"
	RefundStatusSucceeded = "succeeded"
)

type RefundRequest struct {
	PaymentID   string  `json:"payment_id"`  //Идентификатор платежа
	Amount      Amount  `json:"amount"`      //Сумма, которую нужно вернуть пользователю.
	Description string  `json:"description"` //Комментарий к возврату, основание для возврата денег пользователю
	Receipt     Receipt `json:"receipt"`     //Данные для формирования чека в онлайн-кассе (для соблюдения 54-ФЗ). Необходимо указать что-то одно — телефон пользователя (phone) или его электронную почту (email)
}

type Refund struct {
	ID                  string `json:"id"`                   //Идентификатор возврата платежа в Яндекс.Кассе
	PaymentID           string `json:"payment_id"`           //Идентификатор платежа
	Status              string `json:"status"`               //Статус возврата платежа. Возможне значения: canceled, succeeded
	CreatedAt           string `json:"created_at"`           //Время создания возврата. Указывается по UTC и передается в формате ISO 8601, например 2017-11-03T11:52:31.827Z
	Amount              Amount `json:"amount"`               //Сумма, возвращенная пользователю
	ReceiptRegistration string `json:"receipt_registration"` //Статус доставки данных для чека в онлайн-кассу (pending, succeeded или canceled). Присутствует, если вы используете решение Яндекс.Кассы для работы по 54-ФЗ
	Description         string `json:"description"`          //Основание для возврата денег пользователю
}

func (k *Kassa) CreateRefund(inputRefund RefundRequest) (*Refund, *Processing, error) {
	const url = "https://payment.yandex.net/api/v3/refunds"

	serializedRefund, err := json.Marshal(inputRefund)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(serializedRefund))
	if err != nil {
		return nil, nil, err
	}

	req.SetBasicAuth(k.IdempotenceKey, k.SecretKey)
	req.Header.Set("Idempotence-Key", k.IdempotenceKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	switch resp.StatusCode {

	case http.StatusOK:
		var refund Refund
		err = json.Unmarshal(body, &refund)
		if err != nil {
			return nil, nil, err
		}
		return &refund, nil, err

	case http.StatusAccepted:
		var proc Processing
		err = json.Unmarshal(body, &proc)
		if err != nil {
			return nil, nil, err
		}

		return nil, &proc, nil

	default:
		var yandexError Error
		err = json.Unmarshal(body, &yandexError)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, &yandexError
	}
}

func (k *Kassa) RefundInfo(refundId string) (*Refund, *Processing, error) {
	url := fmt.Sprintf("https://payment.yandex.net/api/v3/refunds/%s", refundId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.SetBasicAuth(k.IdempotenceKey, k.SecretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	switch resp.StatusCode {

	case http.StatusOK:
		var refund Refund
		err = json.Unmarshal(body, &refund)
		if err != nil {
			return nil, nil, err
		}
		return &refund, nil, err

	case http.StatusAccepted:
		var proc Processing
		err = json.Unmarshal(body, &proc)
		if err != nil {
			return nil, nil, err
		}

		return nil, &proc, nil

	default:
		var yandexError Error
		err = json.Unmarshal(body, &yandexError)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, &yandexError
	}
}
