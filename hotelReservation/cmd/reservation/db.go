package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"hotel_reserve/common"
	"log"
	"strconv"
)

type Reservation struct {
	HotelId      string `bson:"hotelId"`
	CustomerName string `bson:"customerName"`
	InDate       string `bson:"inDate"`
	OutDate      string `bson:"outDate"`
	Number       int    `bson:"number"`
}

type Number struct {
	HotelId string `bson:"hotelId"`
	Number  int    `bson:"numberOfRoom"`
}

func initializeDatabase(monHelper *common.MonitoringHelper, url string) *mgo.Session {
	fmt.Printf("reservation db ip addr = %v\n", url)

	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	// defer session.Close()

	dbStat1, dbStat2 := monHelper.DBStatTool(common.DbStageLoad)

	c := session.DB("reservation-db").C("reservation")
	count, err := dbStat2(common.DbOpScan, func() (int, error) {
		return c.Find(&bson.M{"hotelId": "4"}).Count()
	})
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		err = dbStat1(common.DbOpInsert, func() error {
			return c.Insert(&Reservation{"4", "Alice", "2015-04-09", "2015-04-10", 1})
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	c = session.DB("reservation-db").C("number")
	count, err = dbStat2(common.DbOpScan, func() (int, error) {
		return c.Find(&bson.M{"hotelId": "1"}).Count()
	})
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		err = dbStat1(common.DbOpInsert, func() error {
			return c.Insert(&Number{"1", 200})
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	count, err = dbStat2(common.DbOpScan, func() (int, error) {
		return c.Find(&bson.M{"hotelId": "2"}).Count()
	})
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		err = dbStat1(common.DbOpInsert, func() error {
			return c.Insert(&Number{"2", 200})
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	count, err = dbStat2(common.DbOpScan, func() (int, error) {
		return c.Find(&bson.M{"hotelId": "3"}).Count()
	})
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		err = dbStat1(common.DbOpInsert, func() error {
			return c.Insert(&Number{"3", 200})
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	count, err = dbStat2(common.DbOpScan, func() (int, error) {
		return c.Find(&bson.M{"hotelId": "4"}).Count()
	})
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		err = dbStat1(common.DbOpInsert, func() error {
			return c.Insert(&Number{"4", 200})
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	count, err = dbStat2(common.DbOpScan, func() (int, error) {
		return c.Find(&bson.M{"hotelId": "5"}).Count()
	})
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		err = dbStat1(common.DbOpInsert, func() error {
			return c.Insert(&Number{"5", 200})
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	count, err = dbStat2(common.DbOpScan, func() (int, error) {
		return c.Find(&bson.M{"hotelId": "6"}).Count()
	})
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		err = dbStat1(common.DbOpInsert, func() error {
			return c.Insert(&Number{"6", 200})
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	for i := 7; i <= 80; i++ {
		hotelId := strconv.Itoa(i)
		count, err = dbStat2(common.DbOpScan, func() (int, error) {
			return c.Find(&bson.M{"hotelId": hotelId}).Count()
		})
		if err != nil {
			log.Fatal(err)
		}
		roomNum := 200
		if i%3 == 1 {
			roomNum = 300
		} else if i%3 == 2 {
			roomNum = 250
		}
		if count == 0 {
			err = dbStat1(common.DbOpInsert, func() error {
				return c.Insert(&Number{hotelId, roomNum})
			})

			if err != nil {
				log.Fatal(err)
			}
		}
	}

	err = c.EnsureIndexKey("hotelId")
	if err != nil {
		log.Fatal(err)
	}

	return session
}
