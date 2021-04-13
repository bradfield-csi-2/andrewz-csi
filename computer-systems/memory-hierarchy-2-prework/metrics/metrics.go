package metrics

import (
	"encoding/csv"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

type UserId int
//type UserMap map[UserId]*User
type PaymentsSlice []int
type AgesSlice []int8
type PaymentsIndices []int

type Address struct {
	fullAddress string
	zip         int
}

type DollarAmount struct {
	dollars, cents uint64
}

type Payment struct {
	amount DollarAmount
	time   time.Time
}

type PaymentsData struct {
	paymentsCents		PaymentsSlice
	dollarAmts			[]DollarAmount
	userIds					[]int
	times						[]time.Time
}


type UsersData struct {
	ids						[]int
	names					[]string
	ages					AgesSlice
	addresses			[]Address
	payments			[]PaymentsIndices
	paymentsData	PaymentsData
}

/*
type User struct {
	id       UserId
	name     string
	age      int
	address  Address
	payments []Payment
}
*/

func AverageAge(users *UsersData) float64 {
	
	ages := users.ages
	stop := len(ages)
	stop1 := stop / 4
	stop2 := stop1 * 2
	stop3 := stop1 * 3


	accChan := make(chan int64, 4)
	defer close(accChan)

	go procAccAge(0, stop1, ages, accChan)
	go procAccAge(stop1, stop2, ages, accChan)
	go procAccAge(stop2, stop3, ages, accChan)
	go procAccAge(stop3, stop, ages, accChan)

	var accSum int64 = 0
	count := 0
	for acc := range accChan {
		accSum += acc
		count += 1
		if count == 4 {
			break
		}
	}
	//pcount := float64(stop)
	return float64(accSum) / float64(stop) //pcount
}

func AveragePaymentAmount(users *UsersData) float64 {
	payments := users.paymentsData.paymentsCents
	stop := len(payments)
	stop1 := stop / 4
	stop2 := stop1 * 2
	stop3 := stop1 * 3


	accChan := make(chan int64, 4)
	defer close(accChan)

	go procAccPayMean(0, stop1, payments, accChan)
	go procAccPayMean(stop1, stop2, payments, accChan)
	go procAccPayMean(stop2, stop3, payments, accChan)
	go procAccPayMean(stop3, stop, payments, accChan)
	
	var accSum int64 = 0
	count := 0
	for acc := range accChan {
		accSum += acc
		count += 1
		if count == 4 {
			break
		}
	}
	pcount := float64(stop * 100)
	return float64(accSum) / pcount 
}


func procAccAge (i, stop int, ages AgesSlice, accChan chan<- int64) {
	stop -= 3
	var acc0, acc1, acc2, acc3 int64 = 0,0,0,0
	for ; i < stop; i += 4  {
		acc0 += int64(ages[i])
		acc1 += int64(ages[i+1])
		acc2 += int64(ages[i+2])
		acc3 += int64(ages[i+3])
	}
	ageLen := stop + 3
	for ;i < ageLen; i++ {
		acc0 += int64(ages[i])	
	}
	accChan <- (acc0 + acc1) + (acc2 + acc3)
}



func procAccPayMean (i, stop int, payments PaymentsSlice, accChan chan<- int64) {
	stop -= 3
	var acc0, acc1, acc2, acc3 int64 = 0,0,0,0
	for ; i < stop; i += 4  {
		acc0 += int64(payments[i])
		acc1 += int64(payments[i+1])
		acc2 += int64(payments[i+2])
		acc3 += int64(payments[i+3])
	}
	payLen := stop + 3
	for ;i < payLen; i++ {
		acc0 += int64(payments[i])	
	}
	accChan <- (acc0 + acc1) + (acc2 + acc3)
}

func procAccPayStdDev (i, stop int, payments PaymentsSlice, accChan chan<- float64) {
	stop -= 3
	var acc float64 = 0.0
	for ; i < stop; i += 4  {
		sq0 := int64(payments[i]) * int64(payments[i])
		sq1 := int64(payments[i+1]) * int64(payments[i+1])
		sq2 := int64(payments[i+2]) * int64(payments[i+2])
		sq3 := int64(payments[i+3]) * int64(payments[i+3])
		sqSum := (sq0 + sq1)  + (sq2 + sq3)
		acc += float64(sqSum) / 10000.0
	}
	payLen := stop + 3
	var sqSum int64 = 0
	for ;i < payLen; i++ {
		sqSum += int64(payments[i])	* int64(payments[i])
	}

	acc += float64(sqSum) / 10000.0
	accChan <- acc
}


// Compute the standard deviation of payment amounts
func StdDevPaymentAmount(users *UsersData) float64 {
	payments := users.paymentsData.paymentsCents

	mean := AveragePaymentAmount(users)
	
	stop := len(payments)
	stop1 := stop / 4
	stop2 := stop1 * 2
	stop3 := stop1 * 3


	accChan := make(chan float64, 4)
	defer close(accChan)

	go procAccPayStdDev(0, stop1, payments, accChan)
	go procAccPayStdDev(stop1, stop2, payments, accChan)
	go procAccPayStdDev(stop2, stop3, payments, accChan)
	go procAccPayStdDev(stop3, stop, payments, accChan)

	var accSum float64 = 0
	count := 0
	for acc := range accChan {
		accSum += acc
		count += 1
		if count == 4 {
			break
		}
	}
	eSS := accSum / float64(stop)
	meanSq := mean * mean
	return math.Sqrt(eSS - meanSq)
}

func LoadData() *UsersData {
	f, err := os.Open("users.csv")
	if err != nil {
		log.Fatalln("Unable to read users.csv", err)
	}
	reader := csv.NewReader(f)
	userLines, err := reader.ReadAll()
	if err != nil {
		log.Fatalln("Unable to parse users.csv as csv", err)
	}

	users := new(UsersData)//make(UserMap, len(userLines))
	users.ids = make([]int,0,1000000)
	users.names = make([]string,0,1000000)
	users.ages = make(AgesSlice,0,1000000)
	users.addresses = make([]Address,0,1000000)
	users.payments = make([]PaymentsIndices,0,1000000)


	for _, line := range userLines {
		id, _ := strconv.Atoi(line[0])
		name := line[1]
		age, _ := strconv.Atoi(line[2])
		address := line[3]
		zip, _ := strconv.Atoi(line[3])
		users.ids = append(users.ids, id)
		users.names = append(users.names, name)
		users.ages = append(users.ages, int8(age))
		users.addresses = append(users.addresses, Address{address,zip})
		users.payments = append(users.payments, make(PaymentsIndices,0))
	}


	f, err = os.Open("payments.csv")
	if err != nil {
		log.Fatalln("Unable to read payments.csv", err)
	}
	reader = csv.NewReader(f)
	paymentLines, err := reader.ReadAll()
	if err != nil {
		log.Fatalln("Unable to parse payments.csv as csv", err)
	}


	for payIdx, line := range paymentLines {
		userId, _ := strconv.Atoi(line[2])
		paymentCents, _ := strconv.Atoi(line[0])
		datetime, _ := time.Parse(time.RFC3339, line[1])

		users.paymentsData.paymentsCents = append(users.paymentsData.paymentsCents, paymentCents)
		users.paymentsData.dollarAmts  = append(users.paymentsData.dollarAmts, DollarAmount{uint64(paymentCents / 100), uint64(paymentCents % 100)})
		users.paymentsData.userIds = append(users.paymentsData.userIds, userId)
		users.paymentsData.times = append(users.paymentsData.times,datetime)
		users.payments[userId] = append(users.payments[userId], payIdx)
	}

	return users
}
