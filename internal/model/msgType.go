package model

type MessageContent interface {
	messageType()
}

type FormattedText struct {
	Text string `json:"text"`
}

type MessageText struct {
	Formatted FormattedText `json:"text"`
}

func (m MessageText) messageType() {}

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

type MessagePhoto struct {
	Caption   FormattedText `json:"caption"`
	PhotoInfo Photo         `json:"photo"`
}

func (m MessagePhoto) messageType() {}

type TdDocument struct {
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	File     TdFile `json:"document"`
}

type MessageDocument struct {
	Caption  FormattedText `json:"caption"`
	Document TdDocument    `json:"document"`
}

func (m MessageDocument) messageType() {}

type TdVideo struct {
	Video TdFile `json:"video"`
}

type MessageVideo struct {
	Caption FormattedText `json:"caption"`
	Video   TdVideo       `json:"video"`
}

func (m MessageVideo) messageType() {}
