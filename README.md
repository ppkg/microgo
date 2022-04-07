# microgo
microservices for go

gin http server
```
func main() {
	flag.Parse()
	defer glog.Flush()

	microgo.Init(&sys.Options{
		Name:           "grpc-gin-empty",
		JaegerEndpoint: "http://10.11.32.165:14268/api/traces",
	}).Run(func(r *gin.RouterGroup) {
		r.GET("/test", api.Test)
		r.POST("/test", api.Test)
	})
}
```
grpc server
```
func main() {
	flag.Parse()
	defer glog.Flush()

	//初始化定时任务，先加任务后执行定时任务
	tasks.InitTask()

	microgo.Init(&sys.Options{
		Name:           "order-service",
		JaegerEndpoint: "http://10.11.32.165:14268/api/traces",
	}).Run(func(s *grpc.Server) {
		order.RegisterOrderServer(s, &services.Order{})
	}, order.RegisterOrderHandler)
}

```