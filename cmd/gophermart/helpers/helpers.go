package helpers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/AlekseyKas/gophermart/internal/config"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

func ControlStatus(wg *sync.WaitGroup, ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			defer wg.Done()
			logrus.Info("Stop checking status")
			return
		case <-time.After(time.Second * 2):
			var order storage.Order
			var Ords []storage.Order
			rows, err := storage.DB.Con.Query(storage.DB.Ctx, "SELECT * FROM orders WHERE status = $1 OR status = $2 OR status = $3", "NEW", "PROCESSING", "REGISTERED")
			if err != nil {
				logrus.Error("Error select orders: ", err)
			}
			for rows.Next() {
				err = rows.Scan(&order.OrderID, &order.UserID, &order.Order, &order.Status, &order.Accrual, &order.UploadedAt)
				if err != nil {
					logrus.Error("Error scan orders: ", err)
				}
				Ords = append(Ords, order)

			}
			if len(Ords) != 0 {
				for i := 0; i < len(Ords); i++ {
					client := resty.New()
					resp, err := client.R().
						SetHeader("Content-Type", "application/json").
						Get(config.Arg.SystemAddress + "/api/orders/" + Ords[i].Order)
					logrus.Info("IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII: ", config.Arg.SystemAddress, " :::: ", string(resp.Body()))
					if err != nil {
						logrus.Error(err)
					}
					order := storage.Orders
					err = json.Unmarshal(resp.Body(), &order)
					if err != nil {
						logrus.Error("Error unmarshal order from accrual: ", err)
					}
					//update status &accrual
					if Ords[i].Status != order.Status {
						_, err = storage.DB.Con.Exec(storage.DB.Ctx, "UPDATE orders SET status = $1, accrual = $2  WHERE number = $3;", order.Status, order.Accrual, order.Order)
						if err != nil {
							logrus.Error("Error update accrual: ", err)
						}
					}
					//update balance
					if order.Status == "INVALID" || order.Status == "PROCESSED" {
						var balance float64
						row, err := storage.DB.Con.Query(storage.DB.Ctx, "SELECT balance FROM balance WHERE user_id = $1", Ords[i].UserID)
						if err != nil {
							logrus.Error("Error select balance: ", err)
						}
						for row.Next() {
							err = row.Scan(&balance)
							if err != nil {
								logrus.Error("Error scan orders: ", err)
							}
						}
						balanceRes := balance + order.Accrual
						_, err = storage.DB.Con.Exec(storage.DB.Ctx, "UPDATE balance SET balance = $1 WHERE user_id = $2;", balanceRes, Ords[i].UserID)

						if err != nil {
							logrus.Error("Error update balance: ", err)
						}
					}

				}
			}
		}
	}

	//for test
	// client := resty.New()
	// _, err := client.R().
	// 	SetHeader("Content-Type", "application/json").
	// 	SetBody(number).
	// 	Post("http://" + config.Arg.SystemAddress + "/api/orders")
	// if err != nil {
	// 	logrus.Error(err)
	// }
	// //for test
	// // client = resty.New()
	// var status string
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		logrus.Info("Check status down")
	// 		defer wg.Done()
	// 		return
	// 	case <-time.After(time.Second * 2):
	// 		defer wg.Done()
	// 		resp, err := client.R().
	// 			SetHeader("Content-Type", "application/json").
	// 			Get("http://" + config.Arg.SystemAddress + "/api/orders/" + string(number))
	// 		if err != nil {
	// 			logrus.Error(err)
	// 		}
	// 		order := storage.Orders
	// 		err = json.Unmarshal(resp.Body(), &order)
	// 		if err != nil {
	// 			logrus.Error("Error unmarshal order from accrual: ", err)
	// 		}
	// 		logrus.Info("sssssssssssssssssssssssss: ", order)
	// 		if order.Status != status {
	// 			err = storage.DB.UpdateOrder(order)
	// 			if err != nil {
	// 				logrus.Error("Error update order in DB: ", err)
	// 			}
	// 			status = order.Status
	// 			if status == "INVALID" || status == "PROCESSED" {
	// 				return
	// 			}
	// 		}
	// 	}
	// }
}
