namespace go push

// 推送服务
service PushService {
	 list<string> funCall(1:i64 callTime, 2:string funCode, 3:map<string, string> paramMap)
}