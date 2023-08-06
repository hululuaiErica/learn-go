package main

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/cache"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stresstest/api/user/gen"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/gormx/callbacks"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/internal/repository"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/internal/repository/dao"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/internal/repository/dao/model"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/internal/service"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"log"
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
	lg, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(lg)

	shadowDB, err := gorm.Open(mysql.Open("root:root@tcp(localhost:11306)/userapp_shadow"))
	if err != nil {
		panic(err)
	}
	shadowDB.AutoMigrate(&model.User{})

	rc := redis.NewClient(&redis.Options{
		Addr:     "localhost:11379",
		Password: "abc",
		DB:       0,
	})

	repo := repository.NewUserRepository(dao.NewUserDAO(shadowDB), cache.NewRedisCache(rc))
	us := service.NewUserService(repo, nil)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context,
		req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		log.Println("我进来了shadow")
		return handler(ctx, req)
	}))
	userapi.RegisterUserServiceServer(server, us)

	l, err := net.Listen("tcp", ":9081")
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
