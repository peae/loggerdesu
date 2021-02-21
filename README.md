# loggerdesu

**使用前先初始化**

> 只初始化一次，不要重复初始化。

Init()中需要传三个参数：appkey, appsecret, clientId. clientId随意填写

```go
	err := loggerdesu.Init("appkey", "appsecret", "clientId")
	if err != nil {
		log.Println(err)
	}
```

**使用**

```go
	loggerdesu.GetSugaredLoggers().Info("Get", "aaa")
	loggerdesu.Sugar.Error("erroooooo")
```
