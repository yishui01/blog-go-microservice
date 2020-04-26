# blog-go-microservice
blog-go-microservice 是一个微服务架构的个人博客。
该项目致力于构建一个分布式博客系统，并不断加入各种微服务组件，以此达到练习微服务架构项目的目的。
本项目大量参考[kratos](https://github.com/go-kratos/kratos)的代码，在此感谢[kratos](https://github.com/go-kratos/kratos)框架带我入门微服务

## 架构图
 ![image](https://file.wuxxin.com/jiagou/a/jiagou.png)

## 仓库目录
| 目录 | 描述 |
| -------- | -------------- |
| app     | 项目应用    |
| app.interface| 聚合服务层 |
| app.service    | 基础服务层   |
| build    |  docker构建、挂载目录  |
| framework | 微服务框架     |

## Quick start
1、启动服务
```bash
make
```
OR
```bash
chmod -R 777 ./build/docker/elasticsearch && docker-compose up;
```
2、访问api
```bash
curl http://localhost:8080/ping
```

## Note
理论上每一个微服务在聚合层都要有一个单独的对外服务，但是由于目前时间原因，将所有微服务的聚合api全部写到了一个main服务中，这样就只有main这一个app对外提供服务了，节约时间，但是耦合度较高，生产环境建议拆分。

## Introduce
### 项目整体介绍
1、用户服务：采用jwt认证，jwt存cookie httponly，另外再在cookie中存一个csrf token，非httponly，用于带在header中。在聚合层 （调用用户服务API）校验用户状态、判断用户权限。

2、文章服务：目前文章查询是直接查询ElasticSearch，详情页单独请求detail接口，从redis中查询文章详情，并使用etcd实现分布式锁，防止缓存击穿

3、诗词服务：需要手动使用 [chinese-poetry-Mysql-Elastic](https://github.com/zuiqiangqishao/chinese-poetry-Mysql-Elastic)  导入2W条诗词到ES中，提供筛选，并随机返回一条res

4、站点信息服务：提供友链、音乐、背景图 等服务

5、使用GoConvey对基础服务层的Dao层进行了测试，覆盖率保证在60%以上（后续会对所有服务的service以及dao层进行测试）

### framework框架介绍
1、通信：所有底层微服务均使用grpc对外提供grpc服务，并用grpc-gateway提供http服务

2、服务注册/发现：微服务启动后使用将ip地址等信息注册到etcd，客户端使用grpc内部resolver（用etcd实现Builder接口）进行服务发现，并使用grpc自带的轮询策略进行客户端负载均衡调用。

3、全链路超时控制：从最外层请求进入时设置context.Timeout, 在grpc server中间件处拦截判断请求剩余时间，将剩余时间和本接口超时时间比对，取小的那个，对ctx设置超时后，再传给handler

4、全链路追踪：最外层使用jaeger生成trace，grpc之间使用client和server两端的grpc_opentracing中间件进行span的传递，http就手动打到header中传递。

5、全链路日志系统：使用zap进行日志库封装，server中间件从context中提取之前注入的traceId，并以此作为本次请求的requestId，对打log的方法进行封装，用该方法打出的每一条log都会带上requestId字段，以此方便elk定位分析每条请求的所有日志。目前主要将日志输出到kafka，logstash从kafka中提取日志再输入到es，通过kibana展示


### 错误处理
    
1、代码全部使用 pkg/errors 的New/Errorf返回错误，用于保存调用堆栈。

2、接受内部代码返回时的err直接透传，不要wrap，防止重复记录堆栈。

3、和标准库/第三方库交互时使用 WithStack/Wrap 来保存调用堆栈。

4、io.EOF这种特殊错误建议不要去Wrap，包了可能会影响其他的代码判断

5、统一在最顶层的调用者处打Log，例如dao层SaveArt调用db出现err时直接 return errors.WithStack(err); 在dao上层的service层Log.Err(...)

6、service层方法直接返回encode或者ecode.Err()，不需要用errors.Wrap包装
#### 错误码处理：

- grpc之间的错误信息统一使用实现 自定义Codes接口 的错误码进行传递
- 这些实现了Codes接口的错误码全部信息统一挂载在grpc的codes.Unknown的detail中，包裹成标准grpc Status传递给rpc调用者
- 实现了Codes接口的错误码有两种:

    第一种：公共错误码，底层是框架已经注册好的ecode，值<=0 （注册定义在framework/pkg/ecode/common_ecode.go）

    ```go   
        framework/pkg/ecode/ecode.go
  
        type Code int
    ```    
    这种主要用于定义公共错误码，全局通用的固定msg错误，比如0 -200、-401、-403、这种状态码在传递时会被统一映射为对应的 http / grpc状态码
    
    第二种：错误码+自定义信息（非固定信息的错误码）
    ```go
    framework/pkg/ecode/status.go
  
     type Status struct {
        s *spb.Status
     }
     ```
     这种是对ecode加上自定义的错误信息，使用ecode.Error/Errorf 来创建
     
     这里的ecode可以是公共错误码，也可以自己手动注册的业务错误码，（使用 ecode.New(code)进行注册，code > 0)
     
     比如表单验证错误：直接使用内置错误码+自己定义的错误信息即可
     ```go
      if err = validate.Struct(req); err != nil {
            err = ecode.Error(ecode.RequestErr, err.Error())
            //当然也可以 err = ecode.RequestErr 不携带任何错误msg提示，让用户自己去反思哪里错了╮(╯▽╰)╭  好吧开玩笑的
            return
      }
     ```
- 传递流程：
    
    => rpc service return ecode &nbsp;&nbsp;

    =>  rpc server 中间件将ecode包裹为grpc.Status{Code:codes.Unknown,message:'自定义的msg / 固定错误码对应的固定msg', Details:上面第二种错误码 Status整个结构体} 
     
    => grpc client 中间件收到response以及err之后，先将err转换为grpc.Status,再尝试从detail中将信息还原为ecode status，转换不了的内部grpc err，也将其转换为ecode错误码，ecode已覆盖了大部分常见grpc 状态码，没有覆盖到的，直接转成 ecode.ServerErr 
    
    => client 调用者函数收到ecode编码的错误
    
    这样就实现了client和server端的ecode交互
    
- 终端http响应
    
    在返回给终端之前（一般是由网关返回给终端），最后负责response的方法收到call方法的ecode后，统一将其归纳到公共错误码ecode中，并将其映射成对应的http statusCode，让客户端收到有限的几种http状态码，而不是各种各样的自定义http状态码。
    
    
        
