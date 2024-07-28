package wire

import "time"

// AlgorithmHashSlots algorithm in routing
const (
	AlgorithmHashSlots = "hashslots"
)

// Command defined data type between client and server
const (
	// login
	CommandLoginSignIn  = "login.signin"
	CommandLoginSignOut = "login.signout"

	// chat
	CommandChatUserTalk  = "chat.user.talk"
	CommandChatGroupTalk = "chat.group.talk"
	CommandChatTalkAck   = "chat.talk.ack"

	// offline
	CommandOfflineIndex   = "chat.offline.index"
	CommandOfflineContent = "chat.offline.content"

	// group manage
	CommandGroupCreate  = "chat.group.create"
	CommandGroupJoin    = "chat.group.join"
	CommandGroupQuit    = "chat.group.quit"
	CommandGroupMembers = "chat.group.members"
	CommandGroupDetail  = "chat.group.detail"
)

// Meta key of a packet
const (
	MetaDestServer   = "dest.server"
	MetaDestChannels = "dest.channels"
)

type Protocol string

const (
	ProtocolTCP       Protocol = "tcp"
	ProtocolUDP       Protocol = "udp"
	ProtocolWebsocket Protocol = "websocket"
)

const (
	SNWGateway = "wgateway"
	SNTGateway = "tgateway"
	SNLogin    = "chat"    //login
	SNChat     = "chat"    //chat
	SNService  = "service" //rpc ser
)

type ServiceID string

type SessionID string

type Magic [4]byte

var (
	MagicLogicPkt = Magic{0xc3, 0x11, 0xa3, 0x65}
	MagicBasicPkt = Magic{0xc3, 0x15, 0xa7, 0x65}
)

const (
	OfflineMessageExpiresIn = time.Hour * 24 * 30
	OfflineSyncIndexCount   = 3000
	OfflineMessageStoreDays = 30 //days
)

const (
	MessageTypeText  = 1
	MessageTypeImage = 2
	MessageTypeVoice = 3
	MessageTypeVideo = 4
)
