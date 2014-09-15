namespace go message

// 离线消息服务
service MessageService {
	void SavePrivateMessage(1:string key, 2:string message, 3:i64 msgId, 4:i64 expire)
}