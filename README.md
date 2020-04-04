# blog-go-microservice

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
    
    
        
