package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

func main() {
	tasks, err := GetTasks()
	if err != nil {
		fmt.Println("Error getting tasks:", err)
		return
	}

	var results []Rectangle
	for _, circles := range tasks {
		boundingBox := CalculateBoundingBox(circles)
		results = append(results, boundingBox)
	}

	if err := CheckResults(results); err != nil {
		fmt.Println("Error checking results:", err)
		return
	}
}

const (
	apiKey = "wGGKtE45kMrF/q8TBbOcWxcy0n5Qm/y//XeNj4r/WDGb8eZK6dG5K+xdqRdrPb8CP7rzdcdR4jSm9AJ7EsIOmA=="
	apiURL = "http://contest.elecard.ru/api"
)

// Circle представляет круг с центром (x, y) и радиусом radius.
type Circle struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Radius float64 `json:"radius"`
}

// Rectangle представляет прямоугольник с двумя точками.
type Rectangle struct {
	LeftBottom Point `json:"left_bottom"`
	RightTop   Point `json:"right_top"`
}

// Point представляет точку с координатами (x, y).
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// APIRequest представляет структуру запроса к API.
type APIRequest struct {
	Key    string      `json:"key"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

// APIResponse представляет структуру ответа от API.
type APIResponse struct {
	Result interface{} `json:"result"`
	Error  *APIError   `json:"error"`
}

// APIError представляет структуру для хранения ошибок от API.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GetTasks получает тестовые случаи с сервера.
func GetTasks() ([][]Circle, error) {
	reqBody, err := json.Marshal(APIRequest{
		Key:    apiKey,
		Method: "GetTasks",
		Params: nil,
	})
	if err != nil {
		return nil, err
	}

	//Отправляем запросик на сервер и получаем ответ в resp в формате json
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//Декодируем ответ от сервера
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("API error: %v", apiResp.Error.Message)
	}

	tasks, ok := apiResp.Result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	// Перебираем result и добавляем в двумерный срез кружочки из ответа
	var result [][]Circle
	for _, task := range tasks {
		var circles []Circle
		taskData, _ := json.Marshal(task)
		if err := json.Unmarshal(taskData, &circles); err != nil {
			return nil, err
		}
		result = append(result, circles)
	}
	return result, nil
}

// CalculateBoundingBox вычисляет минимальный прямоугольник
func CalculateBoundingBox(circles []Circle) Rectangle {
	if len(circles) == 0 {
		return Rectangle{}
	}

	leftBottomX := math.MaxFloat64
	leftBottomY := math.MaxFloat64
	rightTopX := -math.MaxFloat64
	rightTopY := -math.MaxFloat64

	// проходим по паре кругов и ищем у них максимальные и минимальные точки
	for _, circle := range circles {
		leftBottomX = math.Min(leftBottomX, circle.X-circle.Radius)
		leftBottomY = math.Min(leftBottomY, circle.Y-circle.Radius)
		rightTopX = math.Max(rightTopX, circle.X+circle.Radius)
		rightTopY = math.Max(rightTopY, circle.Y+circle.Radius)
	}

	// Возвращаем точки нашего минимального прямоугольника
	return Rectangle{
		LeftBottom: Point{X: leftBottomX, Y: leftBottomY},
		RightTop:   Point{X: rightTopX, Y: rightTopY},
	}
}

// Отправляем наш ответ на проверку серверу
func CheckResults(results []Rectangle) error {
	var params []Rectangle
	for _, result := range results {
		params = append(params, result)
	}

	reqBody, err := json.Marshal(APIRequest{
		Key:    apiKey,
		Method: "CheckResults",
		Params: params,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}

	if apiResp.Error != nil {
		return fmt.Errorf("API error: %v", apiResp.Error.Message)
	}

	resultsCheck, ok := apiResp.Result.([]interface{})
	if !ok {
		return fmt.Errorf("invalid response format for results check")
	}

	for i, result := range resultsCheck {
		if success, ok := result.(bool); ok {
			fmt.Printf("Test %d passed: %v\n", i+1, success)
		}
	}

	return nil
}
