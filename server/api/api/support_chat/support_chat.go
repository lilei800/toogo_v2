// =================================================================================
// Code generated and maintained by GoFrame CLI tool style. (Manually added)
// =================================================================================

package support_chat

import (
	"context"

	"hotgo/api/api/support_chat/v1"
)

type ISupportChatV1 interface {
	Start(ctx context.Context, req *v1.StartReq) (res *v1.StartRes, err error)
	Send(ctx context.Context, req *v1.SendReq) (res *v1.SendRes, err error)
	MessageList(ctx context.Context, req *v1.MessageListReq) (res *v1.MessageListRes, err error)
}


