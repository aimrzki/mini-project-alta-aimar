package controllers

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sashabaranov/go-openai"
	"net/http"
	"os"
	"strings"
)

type HotelResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

type HotelUsecase interface {
	RecommendHotel(userInput, openAIKey string) (string, error)
}

type hotelUsecase struct{}

func NewHotelUsecase() HotelUsecase {
	return &hotelUsecase{}
}

func (uc *hotelUsecase) RecommendHotel(userInput, openAIKey string) (string, error) {
	ctx := context.Background()
	client := openai.NewClient(openAIKey)
	model := openai.GPT3Dot5Turbo
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Halo, perkenalkan saya sistem untuk rekomendasi hotel",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userInput,
		},
	}

	resp, err := uc.getCompletionFromMessages(ctx, client, messages, model)
	if err != nil {
		return "", err
	}
	answer := resp.Choices[0].Message.Content
	return answer, nil
}

func (uc *hotelUsecase) getCompletionFromMessages(
	ctx context.Context,
	client *openai.Client,
	messages []openai.ChatCompletionMessage,
	model string,
) (openai.ChatCompletionResponse, error) {
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)
	return resp, err
}

func RecommendHotel(c echo.Context, hotelUsecase HotelUsecase) error {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": true, "message": "Authorization token is missing"})
	}

	authParts := strings.SplitN(tokenString, " ", 2)
	if len(authParts) != 2 || authParts[0] != "Bearer" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": true, "message": "Invalid token format"})
	}

	tokenString = authParts[1]

	var requestData map[string]interface{}
	err := c.Bind(&requestData)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": true, "message": "Invalid JSON format"})
	}

	userInput, ok := requestData["message"].(string)
	if !ok || userInput == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": true, "message": "Invalid or missing 'message' in the request"})
	}

	userInput = fmt.Sprintf("Rekomendasi hotel: %s", userInput)

	answer, err := hotelUsecase.RecommendHotel(userInput, os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		errorMessage := "Failed to generate hotel recommendations"
		if strings.Contains(err.Error(), "rate limits exceeded") {
			errorMessage = "Rate limits exceeded. Please try again later."
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": true, "message": errorMessage})
	}

	responseData := HotelResponse{
		Status: "success",
		Data:   answer,
	}

	return c.JSON(http.StatusOK, responseData)
}
