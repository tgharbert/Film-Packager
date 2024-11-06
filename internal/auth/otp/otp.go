package otp

import (
	"crypto/rand"
	"fmt"
	"time"
)

type OTPData struct {
	Code string
	ExpiresAt time.Time
}

func GenerateOTP() string {
	var otp[4]byte
	_, err := rand.Read(otp[:])
	if err != nil {
		panic("error generating OTP: " + err.Error())
	}
	return fmt.Sprintf("%06d", otp)
}

func NewOTP() OTPData {
	return OTPData {
		Code: GenerateOTP(),
		ExpiresAt: time.Now().Add(5 * time.Minute), // 5 minute lifespan
	}
}