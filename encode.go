package main

import "errors"

var ErrZero = errors.New("zero")

type b struct{}

func NewCoder() Base62Coder {
	return &b{}
}

const base62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func (b) Encode(num uint64) (string, error) {
	if num <= 0 {
		return "", ErrZero
	}

	var result []byte
	for num > 0 {
		result = append(result, base62[num%62])
		num /= 62
	}
	return string(result), nil
}

func (b) Decode(str string) (uint64, error) {
	var result uint64
	for _, v := range str {
		result = result*62 + uint64(v)
	}
	return result, nil
}
