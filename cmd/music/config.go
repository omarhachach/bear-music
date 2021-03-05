package main

import "github.com/omarhachach/bear"

type Config struct {
	*bear.Config
	APIKey string `json:"api_key"`
}
