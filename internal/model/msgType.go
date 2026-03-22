package model

type MessageContent interface {
	messageType()
}

type FormattedText struct {
	Text string `json:"text"`
}

type MsgCntText struct {
	Formatted FormattedText `json:"text"`
}

func (m MsgCntText) messageType() {}

type LocalFile struct {
	Path string `json:"path"`
}

type TdFile struct {
	Id    int32     `json:"id"`
	Local LocalFile `json:"local"`
}

type PhotoSize struct {
	Type string `json:"type"`
	File TdFile `json:"photo"`
}

type Photo struct {
	PhotoSizes []PhotoSize `json:"sizes"`
}

type MsgCntPhoto struct {
	Caption   FormattedText `json:"caption"`
	PhotoInfo Photo         `json:"photo"`
}

func (m MsgCntPhoto) messageType() {}

type TdDocument struct {
	File TdFile `json:"document"`
}

type MsgCntDoc struct {
	Caption  FormattedText `json:"caption"`
	Document TdDocument    `json:"document"`
}

func (m MsgCntDoc) messageType() {}

type TdVideo struct {
	Video TdFile `json:"video"`
}

type MsgCntVideo struct {
	Caption FormattedText `json:"caption"`
	Video   TdVideo       `json:"video"`
}

func (m MsgCntVideo) messageType() {}
