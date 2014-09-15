namespace go push

// 测试服务
service PushService {
 // 发起远程调用
 list<string> funCall(1:i64 callTime, 2:string funCode, 3:map<string, string> paramMap)
}