/*
 * @Description: In User Settings Edit
 * @Author: your name
 * @Date: 2019-09-27 08:47:02
 * @LastEditTime: 2019-09-27 08:47:02
 * @LastEditors: your name
 */
package loggerdesu

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	// "github.com/gocql/gocql"
	"github.com/peae/loggerdesu/pb"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	"gopkg.in/natefinch/lumberjack.v2"

	stan "github.com/nats-io/go-nats-streaming"
)

//SerName 服务名
var SerName string
var Appkey string
var Appsercet string

// func Getkey(appkey string, appsercet string) error {
// 	var tableName string
// 	cluster := gocql.NewCluster("118.24.5.107")
// 	cluster.Keyspace = "test"
// 	cluster.Consistency = 1
// 	session, _ := cluster.CreateSession()
// 	//设置连接池的数量,默认是2个（针对每一个host,都建立起NumConns个连接）
// 	cluster.NumConns = 3

// 	iter := session.Query(`SELECT table_name FROM congruentrelationship WHERE app_key = ? AND app_secret = ? allow filtering`, appkey, appsercet).Iter()
// 	for iter.Scan(&tableName) {
// 	}

// 	if err := iter.Close(); err != nil {
// 		log.Println(err, "查询出错")
// 		return err
// 	}
// 	fmt.Print(tableName)
// 	SerName = tableName

// 	return nil
// }

const (
	_oddNumberErrMsg    = "Ignored key without a value."
	_nonStringKeyErrMsg = "Ignored key-value pairs with non-string keys."
	channel             = "order-notification"
	event               = "OrderCreated"
	aggregate           = "order"
	grpcUri             = "118.24.5.107:50051"

	clusterID  = "test-cluster"
	clientID   = "order-query-store3"
	durableID  = "store-durable"
	queueGroup = "order-query-store-group"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger
var Sc stan.Conn

func createOrder(Tablename string, Time int64, Package string, Funcname string, Line string, Text string, Type string, Tagname string) {
	log.Println("-=============================ooo3")
	var order pb.Order

	aggregateID := uuid.NewV4().String()
	order.OrderId = aggregateID
	order.Tablename = Tablename
	order.Time = Time
	order.Text = Text
	order.AppKey = Appkey
	order.AppSercet = Appsercet
	order.Package = Package
	order.Funcname = Funcname
	order.Line = Line
	order.Msg = Text
	order.Type = Type
	order.Tagname = Tagname
	//设置服务名
	//设置时间
	//设置日志内容
	// order.OrderItems[0].Code = Code
	// order.OrderItems[0].Name = Name
	// order.OrderItems[0].UnitPrice = UnitPrice
	// order.OrderItems[0].Quantity = Quantity

	//调用发送消息
	// err := createOrderRPC(order)
	// if err != nil {
	// 	fmt.Println("grpc错误")
	// 	log.Print(err)
	// 	return
	// }

	//调用nats发送消息
	err := createOrderNats(order)
	log.Println("-=============================ooo9")

	if err != nil {
		log.Println("-=============================ooo10")
		fmt.Println("nats错误")
		log.Print(err)
		return
	}
	log.Println("-=============================ooo11111")
}

func createOrderNats(order pb.Order) error {

	log.Println("-=============================ooo1")

	// if err != nil {
	// 	Sugar.Info(err)
	// } else {
	// 	Sugar.Info("1111")
	// }

	orderJSON, _ := json.Marshal(order)
	log.Println("-=============================ooo7")
	fmt.Println(string(orderJSON))
	log.Println("-=============================ooo8")
	err := Sc.Publish(channel, orderJSON)
	log.Println("-=============================oooo2")
	return err
	// if err != nil {

	// 	log.Println("-=============================oooo444445")
	// 	return errors.Wrap(err, "Error from nats server")
	// } else {
	// 	return nil
	// }

}

//发送grpc消息
func createOrderRPC(order pb.Order) error {
	//连接grpc服务器
	conn, err := grpc.Dial(grpcUri, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}

	defer conn.Close()
	client := pb.NewEventStoreClient(conn)
	orderJSON, _ := json.Marshal(order)

	event := &pb.Event{
		EventId:       uuid.NewV4().String(),
		EventType:     event,
		AggregateId:   order.OrderId,
		AggregateType: aggregate,
		EventData:     string(orderJSON),
		Channel:       channel,
	}

	resp, err := client.CreateEvent(context.Background(), event)
	if err != nil {
		return errors.Wrap(err, "Error from RPC server")
	}
	if resp.IsSuccess {
		return nil
	} else {
		return errors.Wrap(err, "Error from RPC server")
	}

}

func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func Init(appkey string, appserct string, clientId2 string) error {
	//根据appkey和appsercet获取tablename
	//err := zap.Getkey("bqdefklopgp1hg11ofh0", "bqdefklopgp1hg11ofhg")

	//连接nats服务器
	sc, err := stan.Connect(clusterID, clientId2, stan.NatsURL("nats://118.24.5.107:4222"))
	if err != nil {
		log.Fatal(err)
		return err
	}

	Sc = sc
	Appkey = appkey
	Appsercet = appserct

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "service.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(NewEncoderConfig()),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout),
			w),
		zap.DebugLevel,
	)
	Logger = zap.New(core, zap.AddCaller())
	Sugar = Logger.Sugar()
	return nil
}

type SugaredLoggers struct {
}

func GetSugaredLoggers() *SugaredLoggers {
	log.Println("-=============================ooo4")
	return &SugaredLoggers{}
}

// Info uses fmt.Sprint to construct and log a message.//监听grpc,将信息插入到数据库
func (s *SugaredLoggers) Info(Tagname string, args ...interface{}) {
	log.Println("-=============================ooo56")
	//循环遍历不定参数
	var c string
	for _, item := range args {

		if value, ok := item.(error); ok {
			//知识err类型
			c = c + value.Error()
		} else if value, ok := item.(string); ok {
			c = c + value
		} else if value, ok := item.(int); ok {
			//将int类型转换为string类型
			str1 := strconv.Itoa(value)
			c = c + str1
		}

	}

	t1 := time.Now().Unix() //1564552562
	//发送订单信息

	pc, _, Line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	Funcname := strings.Split(f.Name(), ".")[1]
	Package := strings.Split(f.Name(), ".")[0]

	createOrder(SerName, t1, Package, Funcname, strconv.Itoa(Line), c, "info", Tagname)
	log.Println("-=============================ooo22222222222")
	//请求网关服务
	//如果成功就打印日志，失败就打印连接失败

	// Sugar.Info(args)
	log.Println("-=============================ooo2222222222222324324324")
}

// Warn uses fmt.Sprint to construct and log a message.
func (s *SugaredLoggers) Warn(Tagname string, args ...interface{}) {
	//循环遍历不定参数
	var c string
	for _, item := range args {

		if value, ok := item.(error); ok {
			//知识err类型
			c = c + value.Error()
		} else if value, ok := item.(string); ok {
			c = c + value
		} else if value, ok := item.(int); ok {
			//将int类型转换为string类型
			str1 := strconv.Itoa(value)
			c = c + str1
		}

	}
	fmt.Println(c)
	t1 := time.Now().Unix() //1564552562
	//发送订单信息
	pc, _, Line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	Funcname := strings.Split(f.Name(), ".")[1]
	Package := strings.Split(f.Name(), ".")[0]

	createOrder(SerName, t1, Package, Funcname, strconv.Itoa(Line), c, "warn", Tagname)
	// Sugar.Warn(args)
}

// Error uses fmt.Sprint to construct and log a message.
func (s *SugaredLoggers) Error(Tagname string, args ...interface{}) {
	//循环遍历不定参数
	var c string
	for _, item := range args {

		if value, ok := item.(error); ok {
			//知识err类型
			c = c + value.Error()
		} else if value, ok := item.(string); ok {
			c = c + value
		} else if value, ok := item.(int); ok {
			//将int类型转换为string类型
			str1 := strconv.Itoa(value)
			c = c + str1
		}

	}
	fmt.Println(c)
	t1 := time.Now().Unix() //1564552562
	//发送订单信息
	pc, _, Line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	Funcname := strings.Split(f.Name(), ".")[1]
	Package := strings.Split(f.Name(), ".")[0]

	createOrder(SerName, t1, Package, Funcname, strconv.Itoa(Line), c, "error", Tagname)
	// Sugar.Error(args)
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (s *SugaredLoggers) DPanic(Tagname string, args ...interface{}) {
	//循环遍历不定参数
	var c string
	for _, item := range args {

		if value, ok := item.(error); ok {
			//知识err类型
			c = c + value.Error()
		} else if value, ok := item.(string); ok {
			c = c + value
		} else if value, ok := item.(int); ok {
			//将int类型转换为string类型
			str1 := strconv.Itoa(value)
			c = c + str1
		}

	}
	fmt.Println(c)
	t1 := time.Now().Unix() //1564552562
	//发送订单信息
	pc, _, Line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	Funcname := strings.Split(f.Name(), ".")[1]
	Package := strings.Split(f.Name(), ".")[0]

	createOrder(SerName, t1, Package, Funcname, strconv.Itoa(Line), c, "dpanic", Tagname)
	// Sugar.DPanic(args)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (s *SugaredLoggers) Panic(Tagname string, args ...interface{}) {
	//循环遍历不定参数
	var c string
	for _, item := range args {

		if value, ok := item.(error); ok {
			//知识err类型
			c = c + value.Error()
		} else if value, ok := item.(string); ok {
			c = c + value
		} else if value, ok := item.(int); ok {
			//将int类型转换为string类型
			str1 := strconv.Itoa(value)
			c = c + str1
		}

	}
	fmt.Println(c)
	t1 := time.Now().Unix() //1564552562
	//发送订单信息
	pc, _, Line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	Funcname := strings.Split(f.Name(), ".")[1]
	Package := strings.Split(f.Name(), ".")[0]

	createOrder(SerName, t1, Package, Funcname, strconv.Itoa(Line), c, "panic", Tagname)
	// Sugar.Panic(args)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (s *SugaredLoggers) Fatal(Tagname string, args ...interface{}) {
	//循环遍历不定参数
	var c string
	for _, item := range args {

		if value, ok := item.(error); ok {
			//知识err类型
			c = c + value.Error()
		} else if value, ok := item.(string); ok {
			c = c + value
		} else if value, ok := item.(int); ok {
			//将int类型转换为string类型
			str1 := strconv.Itoa(value)
			c = c + str1
		}

	}
	fmt.Println(c)
	t1 := time.Now().Unix() //1564552562
	//发送订单信息
	pc, _, Line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	Funcname := strings.Split(f.Name(), ".")[1]
	Package := strings.Split(f.Name(), ".")[0]

	createOrder(SerName, t1, Package, Funcname, strconv.Itoa(Line), c, "fatal", Tagname)
	// Sugar.Fatal(args)
}
