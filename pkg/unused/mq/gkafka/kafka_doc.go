package gkafka

/*
	一、生产者(Producer)常用配置
	配置项						默认值								需要调整的场景
	Producer.RequiredAcks	WaitForLocal(折中) 		要求高可靠性:设为 WaitForAll(所有副本确认)	追求吞吐量:设为 NoResponse(不等待确认)
	Producer.Retry.Max				3				网络不稳定时增加重试次数
	Producer.Return.Successes	false	 			同步生产者必须设为 true(否则无法获取发送结果)
	Producer.Partitioner	HashPartitioner	 		需要均匀分发消息:NewRoundRobinPartitioner  按Key分区:NewHashPartitioner
	Producer.Flush.Bytes	0(不启用)	 			批量发送提升吞吐量:设为 1e6(1MB)
	Producer.Flush.Frequency	0(不启用)	 		定时批量发送:设为 500ms(结合 Flush.Bytes 使用)
	Producer.MaxMessageBytes	1000000 (1MB)	 	发送大消息时调高(需匹配 Broker 的 message.max.bytes)
	二、消费者(Consumer)常用配置
	配置项								默认值								需要调整的场景
	Consumer.Group.Rebalance.Strategy	BalanceStrategyRange	 		需要更均衡的分区分配:BalanceStrategyRoundRobin
	Consumer.Offsets.Initial			OffsetNewest	 				首次消费时从最早的消息开始:OffsetOldest
	Consumer.Fetch.Min						1(字节)	 					提高吞吐量:设为 1024(1KB，减少频繁请求)
	Consumer.Fetch.Default				1024 * 1024 (1MB)	 			根据消息大小调整(避免频繁抓取小消息或单次抓取过大)
	Consumer.MaxWaitTime					250ms	 					控制消费者等待新消息的最大时间(平衡延迟和吞吐量)
	Consumer.Group.Session.Timeout			10s	 						消费者组心跳超时(网络延迟高时适当调大，如 30s)
	三、通用配置(生产者和消费者共用)
	配置项						默认值								需要调整的场景
	Net.DialTimeout				30s	 								网络延迟高时调大(如 60s)	快速失败场景调小(如 5s)
	Net.ReadTimeout				30s	 								处理大消息时调大(需匹配 Consumer.Fetch.Max)
	Net.WriteTimeout			30s									高负载生产者调大(避免超时)
	Metadata.Retry.Max			3	 								Broker 不稳定时增加重试次数(如 5)
	Metadata.RefreshFrequency	10m	 								Broker 拓扑频繁变化时调小(如 1m)
	ClientID					sarama	 							建议设置为有业务意义的标识(便于监控和日志追踪)
	四、安全相关配置(如需)
	配置项						默认值								 需要调整的场景
	Net.SASL.Enable				false	 							启用 SASL 认证(如 PLAIN/SCRAM)
	Net.SASL.User				""						 			设置 Kafka 用户名
	Net.SASL.Password			""	 								设置 Kafka 密码
	Net.TLS.Enable				false	 							启用 TLS 加密通信
*/
