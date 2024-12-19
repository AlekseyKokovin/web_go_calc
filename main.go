package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateResult struct {
	Result float64 `json:"result"`
}
type CalculateError struct {
	Error string `json:"error"`
}

func main() {
	http.HandleFunc("/api/v1/calculate", CalculateHandler)
	fmt.Println("Сервер запущен на порту :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		code := 500
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(CalculateError{Error: "Internal server error"})
		return
	}

	var req CalculateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		code := 500
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(CalculateError{Error: "Internal server error"})
		return
	}

	result, err := Calc(req.Expression)
	if err != nil {
		code := 422
		if strings.HasPrefix(err.Error(), "Internal server error") {
			code = 500
		}
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(CalculateError{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CalculateResult{Result: result})
}

func Calc(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	allowed := "+-*/()0123456789"
	for i := 0; i < len(expression); i++ {
		if !strings.Contains(allowed, string(expression[i])) {
			return 0, fmt.Errorf("Expression is not valid")
		}
	}
	expression_rune := []rune(expression)
	numbers := make(map[int]float64)
	numbers_count := 0
	operations := []string{}
	array_skobochki_start := []int{}
	array_skobochki_end := []int{}
	skobochiki_count := 0

	for i := 0; i < len([]rune(expression)); i++ {
		if numbers_count == 0 && operations == nil && (string(expression[i]) == "/" || string(expression[i]) == "*" || string(expression[i]) == "-" || string(expression[i]) == "+") {
			return 0, fmt.Errorf("Internal server error")
		}
		if string(expression[i]) == ")" || string(expression[i]) == "(" {
			skobochiki_count++
			operations = append(operations, string(expression[i]))
			if string(expression[i]) == "(" {
				array_skobochki_start = append(array_skobochki_start, i)
			}
			if string(expression[i]) == ")" {
				array_skobochki_end = append(array_skobochki_end, i)
			}
		} else if num, err := strconv.Atoi(string(expression[i])); err == nil {
			numbers_count++
			numbers[i] = float64(num)
		} else if string(expression[i]) != " " {
			operations = append(operations, string(expression[i]))
		}
	}
	if skobochiki_count%2 == 1 || len(numbers)-(len(operations)-skobochiki_count) != 1 {
		return 0, fmt.Errorf("Internal server error")
	}
	n := len(array_skobochki_end)

	for i := 0; i < n/2; i++ {
		array_skobochki_end[i], array_skobochki_end[n-i-1] = array_skobochki_end[n-i-1], array_skobochki_end[i]
	}
	if skobochiki_count != 0 {
		for i := len(array_skobochki_start) - 1; i > -1; i-- {
			copiedMap := make(map[int]float64)

			for key, value := range numbers {
				if key >= array_skobochki_start[i] {
					copiedMap[key-array_skobochki_start[i]-1] = value
				}
			}
			result, err := GetResult(string(expression_rune[array_skobochki_start[i]+1:array_skobochki_end[i]]), copiedMap)
			if err == nil {
				for jj := array_skobochki_start[i]; jj <= array_skobochki_end[i]; jj++ {
					if _, err := strconv.Atoi(string(expression_rune[jj])); err == nil {
						delete(numbers, jj)
					}
				}

				numbers[array_skobochki_start[i]] = result
				newArr := []rune{}
				newArr = append(newArr, expression_rune[:array_skobochki_start[i]]...)
				newArr = append(newArr, rune(52))
				if len(expression_rune) >= array_skobochki_end[i]+1 {
					newArr = append(newArr, expression_rune[array_skobochki_end[i]+1:]...)
				}
				expression_rune = newArr
			}
			for key, value := range numbers {
				if key >= array_skobochki_end[i] {
					delete(numbers, key)
					numbers[key-(array_skobochki_end[i]-array_skobochki_start[i])] = value
				}
			}

		}
	}

	copiedMap := make(map[int]float64)

	for key, value := range numbers {
		copiedMap[key] = value
	}

	result, err := GetResult(string(expression_rune), copiedMap)
	if err == nil {
		return float64(result), nil
	}
	return 0, fmt.Errorf("Internal server error")
}

func GetResult(expression string, numbers map[int]float64) (float64, error) {
	order := make(map[int]int)
	max_ord := 0
	for j := 0; j < len(expression); j++ {
		if string(expression[j]) == "*" || string(expression[j]) == "/" {
			max_ord++
			order[j] = max_ord
		} else if string(expression[j]) == "-" || string(expression[j]) == "+" {
			order[j] = 0
		}
	}
	new_order := make(map[int]int)
	for key, value := range order {
		if value == 0 {
			max_ord++
			new_order[max_ord] = key
		} else {
			new_order[value] = key
		}
	}
	for jj := 1; jj <= max_ord; jj++ {
		num1 := float64(numbers[new_order[jj]-1])
		num2 := numbers[new_order[jj]+1]
		result := 0.0
		if string(expression[new_order[jj]]) == "+" {
			result += (num1 + num2)
		} else if string(expression[new_order[jj]]) == "-" {
			result += (num1 - num2)
		} else if string(expression[new_order[jj]]) == "*" {
			result += (num1 * num2)
		} else if string(expression[new_order[jj]]) == "/" {
			if num2 == 0 {
				return 0, fmt.Errorf("Internal server error")
			}
			result += (num1 / num2)
		}
		delete(numbers, new_order[jj]-1)
		delete(numbers, new_order[jj]+1)
		newArr := []rune{}
		newArr = append(newArr, []rune(expression[:new_order[jj]-1])...)
		newArr = append(newArr, rune(52))
		if len(expression) >= new_order[jj]+2 {
			newArr = append(newArr, []rune(expression[new_order[jj]+2:])...)
		}
		expression = string(newArr)
		for i := new_order[jj] + 3; i < len(expression)+2; i++ {
			if value, exists := numbers[i]; exists {
				numbers[i-2] = value
			}
		}
		numbers[new_order[jj]-1] = result
		for i := jj + 1; i <= max_ord; i++ {
			if new_order[i] > new_order[jj]+1 {
				new_order[i] -= 2
			}
		}
	}

	return numbers[0], nil
}
