package models

import "time"

// Channel is a database model for each channel JapanBot is a member of
// Its main purpose is to track the card mode of the channel
type Channel struct {
	UID       int    `model:"uid,primarykey,auto"`
	ChannelID string `model:"channel_id,unique"`
	CardMode  int    `model:"card_mode,0"`
}

// Card is a db model for each Card posted to a chat
type Card struct {
	UID       int       `model:"uid,primarykey,auto"`
	ChannelID string    `model:"channel_id"`
	Phrase    string    `model:"phrase"`
	Timestamp time.Time `model:"timestamp"`
}
