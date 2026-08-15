package connector

// OperationType values from LINE SSE operations.
type OperationType int

const (
	OpSendMessage       OperationType = 25
	OpReceiveMessage    OperationType = 26
	OpSendChatRemoved   OperationType = 41
	OpContactUpdate     OperationType = 49
	OpReadReceipt       OperationType = 55
	OpNotifiedJoinChat  OperationType = 60
	OpNotifiedLeaveChat OperationType = 61
	// OpNotifiedTyping (62) has always been on the wire — every third-party
	// LINE client reference (linepy's OpType table, evex-dev/linejs's
	// line.thrift, nmt3325/line-web's live handler) agrees on 62, and this
	// bridge has simply never dispatched it. Verified against nmt3325/line-web
	// server.js: param1 is the chat mid, param2 is the typer's user mid.
	OpNotifiedTyping         OperationType = 62
	OpUnsendLocal            OperationType = 64
	OpUnsendRemote           OperationType = 65
	OpChatUpdate             OperationType = 121
	OpChatUpdate2            OperationType = 122
	OpInviteIntoChat         OperationType = 123
	OpNotifiedInviteIntoChat OperationType = 124
	OpCancelInvitation       OperationType = 126
	OpDeleteSelfFromChat     OperationType = 127
	OpDeleteOtherFromChat    OperationType = 132
	OpPredefinedReaction     OperationType = 139
	OpReaction               OperationType = 140
	OpBlockContact           OperationType = 6
	OpUnblockContact         OperationType = 7
)

// ContentType values for LINE messages.
type ContentType int

const (
	ContentText             ContentType = 0
	ContentImage            ContentType = 1
	ContentVideo            ContentType = 2
	ContentAudio            ContentType = 3
	ContentSticker          ContentType = 7
	ContentContact          ContentType = 13
	ContentFile             ContentType = 14
	ContentLocation         ContentType = 15
	ContentPostNotification ContentType = 16
	ContentSystem           ContentType = 18
	ContentFlex             ContentType = 22
)

// ToType values for LINE message destinations.
type ToType int

const (
	ToUser  ToType = 0
	ToRoom  ToType = 1
	ToGroup ToType = 2
)
