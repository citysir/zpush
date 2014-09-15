package main

type PushRequest struct {
	apiKey      string
	secretKey   string
	deviceType  uint8 //1: web 2: pc 3:android 4:ios 5:wp
	messageType uint8
	message     string
	tagId       uint32
}
