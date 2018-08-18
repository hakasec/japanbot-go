package models

type DBModelList interface {
	Add(model interface{}) error
	Remove(model interface{}) error
}

type Channel struct {
	UID       int    `model:"uid,primarykey"`
	ChannelID string `model:"channel_id"`
	CardMode  int    `model:"card_mode,0"`
}
