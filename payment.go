package yandexkassa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

/*
	Чтобы принять оплату, необходимо создать объект платежа — PaymentRequest.
	Он содержит всю необходимую информацию для проведения оплаты (сумму, валюту и статус).
	У платежа линейный жизненный цикл, он последовательно переходит из статуса в статус.
*/

const (
	PaymentMethodSberbank      = "sberbank"
	PaymentMethodBankCard      = "bank_card"
	PaymentMethodCash          = "cash"
	PaymentMethodYandexMoney   = "yandex_money"
	PaymentMethodQiwi          = "qiwi"
	PaymentMethodAlfabank      = "alfabank"
	PaymentMethodWebmoney      = "webmoney"
	PaymentMethodApplePay      = "apple_pay"
	PaymentMethodMobileBalance = "mobile_balance"
	PaymentMethodInstallments  = "installments"
)

type PaymentRequest struct {
	Amount            Amount                 `json:"amount"`              //Сумма платежа. Иногда партнеры Яндекс.Кассы берут с пользователя дополнительную комиссию, которая не входит в эту сумму.
	Description       string                 `json:"description"`         //Описание транзакции, которое вы увидите в личном кабинете Яндекс.Кассы, а пользователь — при оплате. Например: «Оплата заказа № 72 для user@yandex.ru».
	Receipt           Receipt                `json:"receipt"`             //Данные для формирования чека в онлайн-кассе (для соблюдения 54-ФЗ). Необходимо указать что-то одно — телефон пользователя (phone) или его электронную почту (email)
	Recipient         Recipient              `json:"recipient"`           //Получатель платежа. Нужен, если вы разделяете потоки платежей в рамках одного аккаунта или создаете платеж в адрес другого аккаунта
	PaymentToken      string                 `json:"payment_token"`       //Одноразовый токен для проведения оплаты, сформированный виджетом Yandex.Checkout.js
	PaymentMethodId   string                 `json:"payment_method_id"`   //Идентификатор сохраненного способа оплаты
	PaymentMethodData PaymentMethodData      `json:"payment_method_data"` //Данные, необходимые для создания способа оплаты (payment_method), которым будет платить пользователь
	Confirmation      Confirmation           `json:"confirmation"`        //Данные, необходимые для инициации выбранного сценария подтверждения платежа пользователем
	SavePaymentMethod bool                   `json:"save_payment_method"` //Сохранение платежных данных (с их помощью можно проводить повторные безакцептные списания). Значение true инициирует создание многоразового payment_method
	Capture           bool                   `json:"capture"`             //Автоматический прием поступившего платежа
	ClientIp          string                 `json:"client_ip"`           //IPv4 или IPv6-адрес пользователя. Если не указан, используется IP-адрес TCP-подключения
	Metadata          map[string]interface{} `json:"metadata"`            //Любые дополнительные данные, которые нужны вам для работы с платежами (например, номер заказа). Передаются в виде набора пар «ключ-значение» и возвращаются в ответе от Яндекс.Кассы. Ограничения: максимум 16 ключей, имя ключа не больше 32 символов, значение ключа не больше 512 символов
	Airline           Airline                `json:"airline"`             //Объект с данными для продажи авиабилетов. Используется только для платежей банковской картой
}

/*
	Подтверждает вашу готовность принять платеж. Платеж можно подтвердить, только если он находится в статусе waiting_for_capture.
	Если платеж подтвержден успешно — значит, оплата прошла, и вы можете выдать товар или оказать услугу пользователю
*/
type PaymentConfirmRequest struct {
	Amount  Amount  `json: "amount"`  //Сумма платежа. Иногда партнеры Яндекс.Кассы берут с пользователя дополнительную комиссию, которая не входит в эту сумму.
	Receipt Receipt `json: "receipt"` //Данные для формирования чека в онлайн-кассе (для соблюдения 54-ФЗ). Необходимо указать что-то одно — телефон пользователя (phone) или его электронную почту (email)
	Airline Airline `json: "airline"` //Объект с данными для продажи авиабилетов. Используется только для платежей банковской картой
}

type Payment struct {
	ID                  string                 `json:"id"`                   //Идентификатор платежа
	Status              string                 `json:"status"`               //Статус платежа. Возможные значения: pending, waiting_for_capture, succeeded и canceled
	Amount              Amount                 `json:"amount"`               //Сумма платежа. Иногда партнеры Яндекс.Кассы берут с пользователя дополнительную комиссию, которая не входит в эту сумму
	Description         string                 `json:"description"`          //Описание транзакции, которое вы увидите в личном кабинете Яндекс.Кассы, а пользователь — при оплате. Например: «Оплата заказа № 72 для user@yandex.ru»
	Recipient           Recipient              `json:"recipient"`            //Получатель платежа. Нужен, если вы разделяете потоки платежей в рамках одного аккаунта или создаете платеж в адрес другого аккаунта
	PaymentMethod       PaymentMethod          `json:"payment_method"`       //Способ оплаты, который был использован для этого платежа
	CapturedAt          string                 `json:"captured_at"`          //Время подтверждения платежа. Указывается по UTC и передается в формате ISO 8601
	CreatedAt           string                 `json:"created_at"`           //Время создания заказа. Указывается по UTC и передается в формате ISO 8601. Пример: 2017-11-03T11:52:31.827Z
	ExpiresAt           string                 `json:"expires_at"`           //Время, до которого вы можете бесплатно отменить или подтвердить платеж. В указанное время платеж в статусе waiting_for_capture будет автоматически отменен. Указывается по UTC и передается в формате ISO 8601. Пример: 2017-11-03T11:52:31.827Z
	Confirmation        ConfirmationResponse   `json:"confirmation"`         //Выбранный способ подтверждения платежа. Присутствует, когда платеж ожидает подтверждения от пользователя
	Test                bool                   `json:"test"`                 //Признак тестовой операции
	RefundedAmount      Amount                 `json:"refunded_amount"`      //Сумма, которая вернулась пользователю. Присутствует, если у этого платежа есть успешные возвраты
	Paid                bool                   `json:"paid"`                 //Признак оплаты заказа
	ReceiptRegistration string                 `json:"receipt_registration"` //Статус доставки данных для чека в онлайн-кассу (pending, succeeded или canceled). Присутствует, если вы используете решение Яндекс.Кассы для работы по 54-ФЗ
	Metadata            map[string]interface{} `json:"metadata"`             //Любые дополнительные данные, которые нужны вам для работы с платежами (например, номер заказа). Передаются в виде набора пар «ключ-значение» и возвращаются в ответе от Яндекс.Кассы. Ограничения: максимум 16 ключей, имя ключа не больше 32 символов, значение ключа не больше 512 символов
}

type Recipient struct {
	GatewayID string `json:"gateway_id"` //Идентификатор шлюза. Используется для разделения потоков платежей в рамках одного аккаунта
}

type PaymentMethodData struct {
	Type string `json:"type"` //Тип объекта (например: bank_card, sberbank)
	Card Card   `json:"card"` //Данные банковской карты (необходимы, если вы собираете данные карты пользователей на своей стороне)
}

type Card struct {
	Number      string `json:"number"`       //Номер банковской карты
	ExpiryYear  string `json:"expiry_year"`  //Срок действия, год, YYYY
	ExpiryMonth string `json:"expiry_month"` //Срок действия, месяц, MM
	CSC         string `json:"csc"`          //Код CVC2 или CVV2, 3 или 4 символа, печатается на обратной стороне карты
	Cardholder  string `json:"cardholder"`   //Имя владельца карты
}

type Confirmation struct {
	Type      string `json:"type"`       //Тип объекта redirect. Сценарий, при котором необходимо отправить пользователя на веб-страницу Яндекс.Кассы для подтверждения платежа (например: redirect)
	Enforce   bool   `json:"enforce"`    //Требование принудительного подтверждения платежа пользователем. Например, требование 3-D Secure при оплате банковской картой (по умолчанию определяется политикой платежной системы)
	ReturnUrl string `json:"return_url"` //URL, на который вернется пользователь после подтверждения или отмены платежа на веб-странице
}

type ConfirmationResponse struct {
	Type            string `json:"type"`             //Тип объекта redirect. Сценарий, при котором необходимо отправить пользователя на веб-страницу Яндекс.Кассы для подтверждения платежа (например: redirect)
	Enforce         bool   `json:"enforce"`          //Требование принудительного подтверждения платежа пользователем. Например, требование 3-D Secure при оплате банковской картой (по умолчанию определяется политикой платежной системы)
	ReturnUrl       string `json:"return_url"`       //URL, на который вернется пользователь после подтверждения или отмены платежа на веб-странице
	ConfirmationUrl string `json:"confirmation_url"` //URL, на который необходимо перенаправить пользователя для подтверждения оплаты
}

type Airline struct {
	BookingReference string      `json:"booking_reference"` //Номер бронирования. Обязателен на этапе создания платежа
	TicketNumber     string      `json:"ticket_number"`     //Уникальный номер билета. Обязателен на этапе подтверждения платежа
	Passengers       []Passenger `json:"passengers"`        //Список пассажиров (обязательно должен быть хотя бы 1 пассажир, максимум — 4)
	Legs             []Leg       `json:"legs"`              //Список перелетов (обязательно должен быть хотя бы 1 перелет, максимум — 4)

}

type Passenger struct {
	FirstName string `json:"first_name"` //Имя пассажира
	LastName  string `json:"last_name"`  //Фамилия пассажира
}

type Leg struct {
	DepartureAirport   string `json:"departure_airport"`   //Код аэропорта вылета по справочнику IATA, например LED
	DestinationAirport string `json:"destination_airport"` //Код аэропорта прилета по справочнику IATA, например AMS
	DepartureDate      string `json:"departure_date"`      //Дата вылета в формате YYYY-MM-DD ISO 8601:2004
}

type PaymentMethod struct {
	Type  string `json:"type"`  //тип объекта
	ID    string `json:"id"`    //Идентификатор способа оплаты
	Saved bool   `json:"saved"` //С помощью сохраненного способа оплаты можно проводить безакцептные списания
	Title string `json:"title"` //Название способа оплаты
	Phone string `json:"phone"` //Телефон пользователя, на который зарегистрирован аккаунт в Сбербанке Онлайн. Необходим для подтверждения оплаты по смс (сценарий подтверждения external). Указывается в формате ITU-T E.164, например 79000000000
}

type Processing struct {
	Type        string `json:"type"`         //тип ошибки (e.g. Processing)
	Description string `json:"description"`  //описание самой ошибки
	RetryAfter  int64  `json: "retry_after"` //Через сколько нужно повторить запрос
}

//
func (k *Kassa) CreatePayment(inputPayment *PaymentRequest) (*Payment, *Processing, error) {
	const url = "https://payment.yandex.net/api/v3/payments"

	serializedPayment, err := json.Marshal(inputPayment)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(serializedPayment))
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
		var payment Payment
		err = json.Unmarshal(body, &payment)
		if err != nil {
			return nil, nil, err
		}

		return &payment, nil, err

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

func (k *Kassa) PaymentInfo(paymentId string) (*Payment, *Processing, error) {
	url := fmt.Sprintf("https://payment.yandex.net/api/v3/payments/%s", paymentId)

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
		var payment Payment
		err = json.Unmarshal(body, &payment)
		if err != nil {
			return nil, nil, err
		}

		return &payment, nil, err

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

func (k *Kassa) PaymentNotification(captureFunction, succeedFunction func(*Kassa, *Payment) error) http.HandlerFunc {
	return func(w http.ResponseWriter, q *http.Request) {
		decoder := json.NewDecoder(q.Body)
		var payment Payment
		err := decoder.Decode(&payment)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			panic(err)
		} else {
			if payment.Status == "waiting_for_capture" {
				if err = captureFunction(k, &payment); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					panic(err)
				}
			} else if payment.Status == "succeeded" {
				if err = succeedFunction(k, &payment); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					panic(err)
				}
			}
		}

	}
}

func (k *Kassa) PaymentConfirm(paymentId string, inputPayment *PaymentRequest) (*Payment, *Processing, error) {
	url := fmt.Sprintf("https://payment.yandex.net/api/v3/payments/%s/capture", paymentId)

	paymentConfirmData := &PaymentConfirmRequest{
		Amount:  inputPayment.Amount,
		Receipt: inputPayment.Receipt,
		Airline: inputPayment.Airline}

	serializedPayment, err := json.Marshal(paymentConfirmData)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(serializedPayment))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set(k.IdempotenceKey, k.SecretKey)
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
		var payment Payment
		err = json.Unmarshal(body, &payment)
		if err != nil {
			return nil, nil, err
		}

		return &payment, nil, err

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

func (k *Kassa) PaymentCancel(paymentId string) (*Payment, *Processing, error) {
	url := fmt.Sprintf("https://payment.yandex.net/api/v3/payments/%s/cancel", paymentId)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set(k.IdempotenceKey, k.SecretKey)
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
		var payment Payment
		err = json.Unmarshal(body, &payment)
		if err != nil {
			return nil, nil, err
		}

		return &payment, nil, err

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
