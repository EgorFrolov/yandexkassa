package main

import (
	"encoding/json"
	"log"
	"net/http"
	"yandexkassa"
)

func PaymentHandler(k *yandexkassa.Kassa) http.HandlerFunc {
	return func(w http.ResponseWriter, q *http.Request) {
		decoder := json.NewDecoder(q.Body)

		var paymentReq yandexkassa.PaymentRequest
		err := decoder.Decode(&paymentReq)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			panic(err)
		}

		payment, processing, err := k.CreatePayment(&paymentReq)

		if err != nil {
			if err, ok := yandexkassa.IsYandexError(err); ok {
				log.Println("yandex error: ", err)
				panic(err)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				panic(err)
			}
		}

		if processing != nil {
			log.Println("PaymentHandler(): processing = ", processing)
		} else {
			log.Println("CreatePaymentHandler(): payment from kassa = ", payment)
		}

	}
}

func cancelPayment(k *yandexkassa.Kassa, paymentId string) {
	log.Println("cancelPayment() called")
	payment, processing, err := k.PaymentCancel(paymentId)
	if err != nil {
		if err, ok := yandexkassa.IsYandexError(err); ok {
			log.Println("yandex error: ", err)
			panic(err)
		} else {
			panic(err)
		}
	}

	if processing != nil {
		log.Println("cancelPayment(): processing = ", processing)
	} else {
		log.Println("cancelPayment(): payment = ", payment)
	}

}

func PaymentInfoHandler(k *yandexkassa.Kassa) http.HandlerFunc {
	return func(w http.ResponseWriter, q *http.Request) {
		payment, processing, err := k.PaymentInfo(q.URL.Query()["ID"][0])
		if err != nil {
			if err, ok := yandexkassa.IsYandexError(err); ok {
				log.Println("yandex error: ", err)
				panic(err)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				panic(err)
			}
		}

		if processing != nil {
			log.Println("PaymentInfoHandler(): processing = ", processing)
		} else {
			log.Println("PaymentInfoHandler(): payment from kassa = ", payment)
		}

	}
}

func RefundInfoHandler(k *yandexkassa.Kassa) http.HandlerFunc {
	return func(w http.ResponseWriter, q *http.Request) {
		refund, processing, err := k.RefundInfo(q.URL.Query()["ID"][0])
		if err != nil {
			if err, ok := yandexkassa.IsYandexError(err); ok {
				log.Println("yandex error: ", err)
				panic(err)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				panic(err)
			}
		}

		if processing != nil {
			log.Println("RefundInfoHandler(): processing = ", processing)
		} else {
			log.Println("RefundInfoHandler(): refund from kassa = ", refund)
		}

	}
}

func RefundHandler(k *yandexkassa.Kassa) http.HandlerFunc {
	return func(w http.ResponseWriter, q *http.Request) {
		decoder := json.NewDecoder(q.Body)
		var refundReq yandexkassa.RefundRequest
		err := decoder.Decode(&refundReq)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			panic(err)
		}

		refund, processing, err := k.CreateRefund(refundReq)
		if err != nil {
			if err, ok := yandexkassa.IsYandexError(err); ok {
				log.Println("yandex error: ", err)
				panic(err)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				panic(err)
			}
		}

		if processing != nil {
			log.Println("RefundHandler(): processing = ", processing)
		} else {
			log.Println("CreatePaymentHandler(): refund from kassa = ", refund)
		}

	}
}

func confirmPayment(k *yandexkassa.Kassa, p *yandexkassa.Payment) error {
	log.Println("confirmPayment() called")
	log.Println("confirmPayment() payment.ID = ", p.ID)

	var paymentReq yandexkassa.PaymentRequest
	payment, processing, err := k.PaymentConfirm(p.ID, &paymentReq)
	if err != nil {
		if err, ok := yandexkassa.IsYandexError(err); ok {
			log.Println("yandex error: ", err)
			return err
		} else {
			return err
		}
	}

	if processing != nil {
		log.Println("confirmPayment(): processing = ", processing)
	} else {
		log.Println("confirmPayment(): payment from PaymentConfirm() = ", payment)
	}
	return nil
	//cancelPayment(payment.ID)
}

func succeedPayment(k *yandexkassa.Kassa, p *yandexkassa.Payment) error {
	log.Println("succeedPayment() called")
	log.Println("succeedPayment() called")
	log.Println("succeedPayment() payment.ID = ", p.ID)
	return nil
}

func main() {
	kassa := &yandexkassa.Kassa{
		ShopID:         12345,
		SecretKey:      "testSecrectKey",
		IdempotenceKey: "testIdempotenceKey"}

	router := http.NewServeMux()
	http.HandleFunc("/create_payment", PaymentHandler(kassa))
	http.HandleFunc("/payment_notify", kassa.PaymentNotification(confirmPayment, succeedPayment))
	http.HandleFunc("/payment_info", PaymentInfoHandler(kassa))
	http.HandleFunc("/create_refund", RefundHandler(kassa))
	http.HandleFunc("/refund_info", RefundInfoHandler(kassa))
	http.ListenAndServe(":4567", router)
}
