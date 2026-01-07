package support_chat

import api "hotgo/api/api/support_chat"

type ControllerV1 struct{}

func NewV1() api.ISupportChatV1 {
	return &ControllerV1{}
}


