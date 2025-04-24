# Scira2api
Transform [scira's web service](https://mcp.scira.ai/)  into an API service, The API supports access in the OpenAI format.

## ✨ Features

- 🔁 **UserId Polling** - userId polling with support for multiple userIds.
- 📝 **Automatic Conversation Management** -  Conversation can be automatically deleted after use
- 🌊 **Streaming Responses** - Get real-time streaming outputs 
- 🌐 **Proxy Support** - Route requests through your preferred proxy
- 🔐 **API Key Authentication** - Secure your API endpoints
- 🔁 **Automatic Retry** - Feature to automatically retry requests when request fail

## 📋 Prerequisites

- Go 1.24+ (for building from source)
- Docker (for containerized deployment)

## 🚀 Deployment Options

### Docker

```bash
docker run -d \
  -p 8080:8080 \
  -e USERIDS=xxx,yyy \
  -e APIKEY=sk-123 \
  -e CHAT_DELETE=true \
  -e HTTP_PROXY=http://127.0.0.1:7890 \
  -e MODELS=gpt-4.1-mini,claude-3-7-sonnet,grok-3-mini,qwen-qwq \
  -e RETRY=3 \
  --name scira2api \
  ghcr.io/coderzoe/scira2api:latest
```

### Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3'
services:
  scira2api:
    image: ghcr.io/coderzoe/scira2api:latest
    container_name: scira2api
    ports:
      - "8080:8080"
    environment:
      - USERIDS=xxx,yyy  # Required
      - APIKEY=sk-123  # Optional
      - CHAT_DELETE=true  # Optional
      - HTTP_PROXY=http://127.0.0.1:7890  # Optional
      - MODELS=gpt-4.1-mini,claude-3-7-sonnet,grok-3-mini,qwen-qwq   # Optional
      - RETRY=3  # Optional
    restart: unless-stopped

```

Then run:

```bash
docker-compose up -d
```

Or：

```bash
# Clone the repository
git clone https://github.com/coderZoe/scira2api.git
cd scira2api
# edit environment
vi docker-compose.yml
./deploy.sh
```



### Direct Deployment

```bash
# Clone the repository
git clone https://github.com/coderZoe/scira2api.git
cd scira2api
cp .env.example .env  
vim .env  
# Build the binary
go build -o scira2api .

./scira2api
```

## ⚙️ Configuration

### ENV Configuration

You can configure `scira2api` using a `.env` file in the application's root directory. If this file exists, it will be used instead of environment variables.

Example `.env`:

```yaml
# Required, separate multiple userIds with English commas
UserIds= xxx,yyy

# Optional, Port. Default: 8080
Port=8080

# Optional, API key for authenticating client requests (e.g., the key entered for openweb-ui requests). If empty, no authentication is required.
ApiKey=sk-xxx

# Optional, Proxy address. Default: No proxy is used.
HTTP_PROXY= http://127.0.0.1:7890

# Optional, List of models, separated by English commas.
Models=gpt-4.1-mini,claude-3-7-sonnet,grok-3-mini,qwen-qwq

# Optional, Number of retry attempts on request failure. 0 or 1 means no retry. Default: 0 (no retry). A different userId will be used for each retry.
Retry=3

# Optional, Whether to delete chat history on the page. Default: false (do not delete).
ChatDelete=true
```

A sample configuration file is provided as `.env.example` in the repository.


## 📝 API Usage

### Authentication

Include your API key in the request header:

```bash
# no need if you not config apiKey
Authorization: Bearer YOUR_API_KEY
```

### Chat Completion

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-4.1-mini",
    "messages": [
      {
        "role": "user",
        "content": "Hello, Man!"
      }
    ],
    "stream": true
  }'
```


## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Made with ❤️ by[coderZoe](https://github.com/coderZoe)