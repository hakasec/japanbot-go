package models

// Channel is a database model for each channel JapanBot is a member of
// Its main purpose is to track the card mode of the channel
type Channel struct {
	UID       int    `model:"uid,primarykey,auto"`
	ChannelID string `model:"channel_id"`
	CardMode  int    `model:"card_mode,0"`
}
