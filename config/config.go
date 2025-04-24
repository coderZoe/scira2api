package config

import (
	"math"
	"os"
	"scira2api/log"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	ApiKey     string
	UserIds    []string
	HttpProxy  string
	Models     []string
	Retry      int
	ChatDelete bool
}

func NewConfig() *Config {
	godotenv.Load()
	port := os.Getenv("Port")
	if port == "" {
		port = "8080"
	}
	userIds := strings.Split(os.Getenv("UserIds"), ",")
	if len(userIds) == 0 {
		log.Fatal("userids is empty")
	}
	proxy := os.Getenv("http_proxy")
	if proxy == "" {
		proxy = os.Getenv("HTTP_PROXY")
	}
	models := strings.Split(os.Getenv("Models"), ",")
	if len(models) == 0 {
		models = []string{"gpt-4.1-mini", "claude-3-7-sonnet", "grok-3-mini", "qwen-qwq"}
	}
	retry := os.Getenv("Retry")
	if retry == "" {
		retry = "1"
	}
	retryInt, err := strconv.Atoi(retry)
	if err != nil {
		log.Fatal("retry is not a number")
	}
	//最少1次
	retryInt = int(math.Max(float64(retryInt), 1))

	chatDelete := os.Getenv("ChatDelete")
	if chatDelete == "" {
		chatDelete = "false"
	}
	chatDeleteBool, err := strconv.ParseBool(chatDelete)
	if err != nil {
		log.Fatal("chatDelete should be true or false")
	}
	return &Config{
		Port:       port,
		ApiKey:     os.Getenv("ApiKey"),
		UserIds:    userIds,
		HttpProxy:  proxy,
		Models:     models,
		Retry:      retryInt,
		ChatDelete: chatDeleteBool,
	}
}
