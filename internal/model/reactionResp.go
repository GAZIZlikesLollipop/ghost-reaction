package model

type CommonReaction struct {
	Type   CommonResp `json:"type"`
	ForVIP bool       `json:"needs_premium"`
}

type CommonReactions struct {
	Top     []CommonReaction `json:"top_reactions"`
	Recent  []CommonReaction `json:"recent_reactions"`
	Popular []CommonReaction `json:"popular_reactions"`
}

type ReactionsResp struct {
	Resp CommonReactions
	Req  []byte
}

type EmojiReactType struct {
	Emoji string `json:"emoji"`
}

type EmojiReaction struct {
	Type EmojiReactType `json:"type"`
}

type EmojiReactions struct {
	Top     []EmojiReaction `json:"top_reactions"`
	Recent  []EmojiReaction `json:"recent_reactions"`
	Popular []EmojiReaction `json:"popular_reactions"`
}
