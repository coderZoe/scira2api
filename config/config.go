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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	userIdsEnv := os.Getenv("USERIDS")
	if userIdsEnv == "" {
		log.Fatal("USERIDS is empty")
	}
	userIds := strings.Split(userIdsEnv, ",")

	proxy := os.Getenv("http_proxy")
	if proxy == "" {
		proxy = os.Getenv("HTTP_PROXY")
	}
	modelsEnv := os.Getenv("MODELS")
	if modelsEnv == "" {
		modelsEnv = "gpt-4.1-mini,claude-3-7-sonnet,grok-3-mini,qwen-qwq"
	}
	models := strings.Split(modelsEnv, ",")
	retry := os.Getenv("RETRY")
	if retry == "" {
		retry = "1"
	}
	retryInt, err := strconv.Atoi(retry)
	if err != nil {
		log.Fatal("RETRY is not a number")
	}
	//最少1次
	retryInt = int(math.Max(float64(retryInt), 1))

	chatDelete := os.Getenv("CHAT_DELETE")
	log.Info("CHAT_DELETE: %s", chatDelete)
	if chatDelete == "" {
		chatDelete = "false"
	}
	log.Info("CHAT_DELETE: %s", chatDelete)
	chatDeleteBool, err := strconv.ParseBool(chatDelete)
	if err != nil {
		log.Fatal("CHAT_DELETE should be true or false")
	}
	return &Config{
		Port:       port,
		ApiKey:     os.Getenv("APIKEY"),
		UserIds:    userIds,
		HttpProxy:  proxy,
		Models:     models,
		Retry:      retryInt,
		ChatDelete: chatDeleteBool,
	}
}
