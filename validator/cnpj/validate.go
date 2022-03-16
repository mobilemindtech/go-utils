package cnpj

import (
	"errors"
	"strconv"
	"strings"	
)

// Valid validates the cpf and return a boolean and the error if any
func IsValid(digits string) (bool, error) {
	return valid(digits)
}

func sanitize(data string) string {
	data = strings.Replace(data, ".", "", -1)
	data = strings.Replace(data, "-", "", -1)
	data = strings.Replace(data, "/", "", -1)
	return data
}

func valid(data string) (bool, error) {
	data = sanitize(data)

	if len(data) != 14 {
		return false, errors.New("Invalid length")
	}

	if strings.Contains(blacklist, data) || !check(data) {
		return false, errors.New("Invalid value")
	}

	return true, nil
}

const blacklist = `00000000000000
11111111111111
22222222222222
33333333333333
44444444444444
55555555555555
66666666666666
77777777777777
88888888888888
99999999999999`

func stringToIntSlice(data string) (res []int) {
	for _, d := range data {
		x, err := strconv.Atoi(string(d))
		if err != nil {
			continue
		}
		res = append(res, x)
	}
	return
}

func check(data string) bool {
	return verify(stringToIntSlice(data), 5, 12) && verify(stringToIntSlice(data), 6, 13)
}

func verify(data []int, j int, n int) bool{
		
	soma := 0		

	for i := 0; i < n; i++ {
		v := data[i]
		soma += v * j
		
		if j == 2 {
			j = 9
		} else {
			j -= 1
		}
	}

	resto := soma % 11

	v := data[n]
	x := 0

	if resto >= 2 {
		x = 11 - resto
	}

	if  v != x {
		return false
	}

	return true
}