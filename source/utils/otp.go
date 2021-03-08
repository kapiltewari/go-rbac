package utils

import (
	"math/rand"
	"strconv"
	"time"
)

//GenerateOTP ...
func GenerateOTP() string {
	//source
	rand.Seed(time.Now().UnixNano())

	//random number
	randomNumber := rand.Uint32()
	randomNumberString := strconv.Itoa(int(randomNumber * rand.Uint32()))

	var otpString string

	if len(randomNumberString) > 8 {
		//otpString length is 8
		otpString = randomNumberString[0:8]
	}

	return otpString
}
