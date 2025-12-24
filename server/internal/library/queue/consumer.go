// Package queue
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package queue

import (
	"context"
	"hotgo/internal/consts"
	"hotgo/utility/simple"
	"sync"
)

// Consumer 消费者接口，实现该接口即可加入到消费队列中
type Consumer interface {
	GetTopic() string                                    // 获取消费主题
	Handle(ctx context.Context, mqMsg MqMsg) (err error) // 处理消息的方法
}

// consumerManager 消费者管理
type consumerManager struct {
	sync.Mutex
	list map[string]Consumer // 维护的消费者列表
}

var consumers = &consumerManager{
	list: make(map[string]Consumer),
}

// RegisterConsumer 注册任务到消费者队列
func RegisterConsumer(cs Consumer) {
	consumers.Lock()
	defer consumers.Unlock()
	topic := cs.GetTopic()
	if _, ok := consumers.list[topic]; ok {
		Logger().Debugf(ctx, "queue.RegisterConsumer topic:%v duplicate registration.", topic)
		return
	}
	consumers.list[topic] = cs
}

// StartConsumersListener 启动所有已注册的消费者监听
func StartConsumersListener(ctx context.Context) {
	for _, c := range consumers.list {
		consumer := c
		// 使用SafeGo避免消费者内部panic导致进程崩溃
		simple.SafeGo(ctx, func(ctx context.Context) {
			consumerListen(ctx, consumer)
		})
	}
}

// consumerListen 消费者监听
func consumerListen(ctx context.Context, job Consumer) {
	var (
		topic  = job.GetTopic()
		c, err = InstanceConsumer()
	)

	if err != nil {
		Logger().Errorf(ctx, "InstanceConsumer %s err:%+v", topic, err)
		return
	}

	// 监听服务关闭事件（方便底层mq驱动在内部cancel/close）
	simple.Event().Register(consts.EventServerClose, func(ctx context.Context, args ...interface{}) {
		Logger().Debugf(ctx, "queue consumer shutdown signal received, topic=%s", topic)
	})

	if listenErr := c.ListenReceiveMsgDo(topic, func(mqMsg MqMsg) {
		err = job.Handle(ctx, mqMsg)

		// if err != nil {
		//	// 遇到错误，重新加入到队列
		//	//queue.Push(topic, mqMsg.Body)
		// }

		// 记录消费队列日志
		ConsumerLog(ctx, topic, mqMsg, err)
	}); listenErr != nil {
		// 运行期错误不应直接退出进程
		Logger().Errorf(ctx, "消费队列：%s 监听失败, err:%+v", topic, listenErr)
	}
}
