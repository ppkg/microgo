# microgo
microservices for go

gin http server
```
func main() {
	flag.Parse()
	defer glog.Flush()

	microgo.Init(&sys.Options{
		Name:           "grpc-gin-empty",
		ConsulAddress: "127.0.0.1:8500",
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
		ConsulAddress: "127.0.0.1:8500",
	}).Run(func(s *grpc.Server) {
		order.RegisterOrderServer(s, &services.Order{})
	}, order.RegisterOrderHandler)
}

```
