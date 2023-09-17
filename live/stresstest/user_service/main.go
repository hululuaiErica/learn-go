package main

import (
	"context"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stresstest/api/user/gen"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/gormx/callbacks"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/gormx/connpool"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/internal/repository"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/internal/repository/dao"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/internal/repository/dao/model"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/internal/service"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/redisx"
	"github.com/Shopify/sarama"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
	"net"
	// rstore "gitee.com/geektime-geekbang/geektime-go/web/session/redis"
	"github.com/go-redis/redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	_ "net/http/pprof"
)

// 这里各种地址都是直接写死的，在真实的环境替换为从配置文件里面读取就可以
// 随便用一个配置框架，大体上都差不多的
func main() {
	//initZipkin()
	// 在 main 函数的入口里面完成所有的依赖组装。
	// 这个部分你可以考虑替换为 google 的 wire 框架，达成依赖注入的效果
	lg, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(lg)
	//cfg := sarama.NewConfig()
	//cfg.Producer.Return.Successes = true
	//producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, cfg)
	//if err != nil {
	//	panic(err)
	//}

	liveDB, err := gorm.Open(mysql.Open("root:root@tcp(localhost:11306)/userapp"),
		&gorm.Config{
			SkipDefaultTransaction: true,
		})
	if err != nil {
		panic(err)
	}
	liveDB.AutoMigrate(&model.User{})

	//group, _ := sarama.NewConsumerGroup([]string{"abc"}, "abc-consumer", cfg)
	//group.Consume(context.Background(), []string{"biz-topic"}, &consumer{})

	shadowDB, err := gorm.Open(mysql.Open("root:root@tcp(localhost:11306)/userapp_shadow"))
	if err != nil {
		panic(err)
	}
	shadowDB.AutoMigrate(&model.User{})

	shadowPool := connpool.NewShadowConnPool(liveDB.ConnPool, shadowDB.ConnPool)

	//readWriteSplitPool := connpool.NewReadWriteSplitPool(liveDB, []*gorm.DB{shadowDB})
	//
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: shadowPool,
	}))
	if err != nil {
		panic(err)
	}

	rc := redis.NewClient(&redis.Options{
		Addr:     "localhost:11379",
		Password: "abc",
		DB:       0,
	})

	shadowRC := redisx.NewShadowCmdablePrefix(rc)

	if err != nil {
		panic(err)
	}
	repo := repository.NewUserRepository(dao.NewUserDAO(db), shadowRC)
	us := service.NewUserService(repo, nil)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context,
		req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			stress := md.Get("stress-test")
			if len(stress) > 0 {
				ctx = context.WithValue(ctx, "stress-test", stress[0])
			}
		}
		return handler(ctx, req)
	}))
	userapi.RegisterUserServiceServer(server, us)

	l, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}
	if err = server.Serve(l); err != nil {
		panic(err)
	}
}

func initDB(liveDB *gorm.DB) {
	liveDB.Callback().Query().Before("*").Register("shadow_query", gormShadowCallback)
	liveDB.Callback().Delete().Before("*").Register("shadow_delete", gormShadowCallback)
	liveDB.Callback().Create().Before("*").Register("shadow_create", gormShadowCallback)
	liveDB.Callback().Update().Before("*").Register("shadow_update", gormShadowCallback)
	//liveDB.Callback().Raw().Before("*").Register("shadow_update", gormShadowCallback)
	//liveDB.Callback().Row().Before("*").Register("shadow_update", gormShadowCallback)

	dstDB, err := gorm.Open(mysql.Open("root:root@tcp(localhost:11306)/userapp_dst"))
	if err != nil {
		panic(err)
	}

	dwcb := callbacks.NewDoubleWriteCallbackBuilder(dstDB).Build()
	liveDB.Callback().Delete().After("gorm:delete").Register("double_write_delete", dwcb)
	liveDB.Callback().Create().After("gorm:create").Register("double_write_create", dwcb)
	liveDB.Callback().Update().After("gorm:update").Register("double_write_update", dwcb)

	bfcb := (&callbacks.BeforeFindBuilder{}).Build()
	liveDB.Callback().Query().Before("gorm:query").Register("before_find", bfcb)
	dstDB.AutoMigrate(&model.User{})
}

type consumer struct {
}

func (c *consumer) Setup(session sarama.ConsumerGroupSession) error {
	//TODO implement me
	panic("implement me")
}

func (c *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	//TODO implement me
	panic("implement me")
}

func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		// 在这里
		// 重新组装你的压测标记
		ctx := context.WithValue(context.Background(),
			string(msg.Headers[0].Key), string(msg.Headers[0].Value))
		bizHandler(ctx)
	}
	return nil
}

func bizHandler(ctx context.Context) {

}

func buildShadowCallback(m map[string]string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		stress := ctx.Value("stress-test")
		// my_users_shadow
		if stress == "true" {
			tblName, ok := m[db.Statement.Table]
			if ok {
				db.Statement.Table = tblName
			} else {
				db.Statement.Table = db.Statement.Table + "_shadow"
			}
		}
	}
}
func gormShadowCallback(db *gorm.DB) {
	ctx := db.Statement.Context
	stress := ctx.Value("stress-test")
	// my_users_shadow
	if stress == "true" {
		if tblName, ok := db.Statement.Model.(interface {
			ShadowTableName() string
		}); ok {
			db.Statement.Table = tblName.ShadowTableName()
		} else {
			db.Statement.Table = db.Statement.Table + "_shadow"
		}
	}
}

//func initZipkin() {
//	exporter, err := zipkin.New(
//		"http://localhost:19411/api/v2/spans",
//		zipkin.WithLogger(log.New(os.Stderr, "userapp", log.Ldate|log.Ltime|log.Llongfile)),
//	)
//	if err != nil {
//		panic(err)
//	}
//	batcher := sdktrace.NewBatchSpanProcessor(exporter)
//	tp := sdktrace.NewTracerProvider(
//		sdktrace.WithSpanProcessor(batcher),
//		sdktrace.WithResource(resource.NewWithAttributes(
//			semconv.SchemaURL,
//			semconv.ServiceNameKey.String("userapp"),
//		)),
//	)
//	otel.SetTracerProvider(tp)
//}
