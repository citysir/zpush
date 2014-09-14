package proto

// The Message struct
type Message struct {
	GroupId uint   `json:"gid"` // group id
	MsgId   int64  `json:"mid"` // message id
	Msg     String `json:"msg"` // message content
}
