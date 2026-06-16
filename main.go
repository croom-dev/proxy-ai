package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var httpClient = &http.Client{}

func main() {

	r := gin.Default()

	r.Any("/openai/*path", func(c *gin.Context) {
		Proxy(
			c,
			"https://api.openai.com",
		)
	})

	r.Any("/gemini/*path", func(c *gin.Context) {
		Proxy(
			c,
			"https://generativelanguage.googleapis.com",
		)
	})

	r.Run(":8888")

	log.Println("AI Proxy running")
}

func Proxy(
	c *gin.Context,
	baseURL string,
) {

	path := c.Param("path")

	target := baseURL + path

	if c.Request.URL.RawQuery != "" {
		target += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequest(
		c.Request.Method,
		target,
		c.Request.Body,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	// copy headers
	for k, v := range c.Request.Header {
		req.Header[k] = v
	}

	// remove client ip leakage
	req.Header.Del("X-Forwarded-For")
	req.Header.Del("X-Real-IP")
	req.Header.Del("Forwarded")

	// provider auth
	// req.Header.Set(
	// 	"Authorization",
	// 	"Bearer "+apiKey,
	// )

	resp, err := httpClient.Do(req)

	if err != nil {
		c.JSON(502, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer resp.Body.Close()

	for k, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(k, value)
		}
	}

	c.Status(resp.StatusCode)

	io.Copy(c.Writer, resp.Body)
}
