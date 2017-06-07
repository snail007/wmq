# WMQ
wrapped  message queue which based on rabbitmq,support http protocol
# Usage:
<pre>
Usage of wmq:
--api-disable                  disable api service
--api-token string             access api token (default "guest")
--data-example                 print example of data-file
--data-file string             which file will store messages (default "message.json")
--fail-wait int                access consumer url  fail and then how many milliseconds to sleep (default 50000)
--ignore-headers stringSlice   these http headers will be ignored when access to consumer's url , 
                                multiple splitted by comma(,)
--level string                 console log level,should be one of debug,info,warn,error (default "debug")
--listen-api string            api service listening port (default "0.0.0.0:3302")
--listen-publish string        publish service listening port (default "0.0.0.0:3303")
--log-access                   access log on or off (default true)
--log-dir string               the directory which store log files (default "log")
--log-level stringSlice        log to file level,multiple splitted by comma(,) (default [info,error,debug])
--log-max-count int            log file max count for rotate to remain (default 3)
--log-max-size int             log file max size(bytes) for rotate (default 102400000)
--mq-host string               which host be used when connect to RabbitMQ (default "127.0.0.1")
--mq-password string           which password be used when connect to RabbitMQ (default "guest")
--mq-port int                  which port be used when connect to RabbitMQ (default 5672)
--mq-prefix string             the queue and exchange default prefix (default "wmq.")
--mq-username string           which username be used when connect to RabbitMQ (default "guest")
--mq-vhost string              which vhost be used when connect to RabbitMQ (default "/")
--realip-header string         the publisher's real ip will be set in this http header when 
                                access to consumer's url (default "X-Forwarded-For")
--version                      show version about current WMQ
</pre>

# Publishing Message
<pre>
note:default publish port is 3303
1.publish a message 
    note:any "post body" and "get parameters" and "http header" was send to 
        "publishing" , them will be the same as when wmq access consumer's URL
    request:
        protocol:http
        method:get or post
        path:/:name?:query_string     //:name is the name of message ,
                                        :query_string is any query string you need
        header:
            Token:string        //message's Token , if not need token ,leave it empty
            RouteKey:string     //the api token is setting in config
    response:
        httpcode:204|500      //204:menas success 500:means fail and output is error info
</pre>

# Management
<pre>
note:default manage port is 3302
1.add a message
    request:
        protocol:http
        method:get
        path:/message/add
        parameters:
            Name:string     //message name,must be unique
            Comment:string  //comment 
            Durable:1|0     //durable or not,1:true,0:false
            IsNeedToken:1|0 //need token or not when publish this kind message,1:true,0:false
            Mode:string     //should be one of fanout,topic,direct
            Token:string    //should be set when IsNeedToken is 1,other leave empty
            api-token:string//the api token is setting in config
            callback:string //callback function name for jsonp call,if no jsonp call ,leave it empty
    response:
        type:json
        column:
            code:1|0    //1 means success , 0 means fail
        example:
            no jsonp:{code:1,data:null} or {code:0,data:"some error"} 
            jsonp:callbackxxx({code:1,data:null}) or callbackxxx({code:0,data:"some error"}) 
2.update a message
    request:
        protocol:http
        method:get
        path:/message/update
        parameters:
            Name:string     //message name
            Comment:string  //comment 
            Durable:1|0     //durable or not,1:true,0:false
            IsNeedToken:1|0 //need token or not when publish this kind message,1:true,0:false
            Mode:string     //should be one of fanout,topic,direct
            Token:string    //should be set when IsNeedToken is 1,other leave empty
            api-token:string//the api token is setting in config
            callback:string //callback function name for jsonp call,if no jsonp call ,leave it empty
    response:
        type:json
        column:
            code:1|0    //1 means success , 0 means fail
        example:
            no jsonp:{code:1,data:null} or {code:0,data:"some error"} 
            jsonp:callbackxxx({code:1,data:null}) or callbackxxx({code:0,data:"some error"}) 
3.delete a message
    request:
        protocol:http
        method:get
        path:/message/delete
        parameters:
            Name:string     //message name
            api-token:string//the api token is setting in config
            callback:string //callback function name for jsonp call,if no jsonp call ,leave it empty
    response:
        type:json
        column:
            code:1|0    //1 means success , 0 means fail
        example:
            no jsonp:{code:1,data:null} or {code:0,data:"some error"} 
            jsonp:callbackxxx({code:1,data:null}) or callbackxxx({code:0,data:"some error"}) 
4.add a consumer
    request:
        protocol:http
        method:get
        path:/consumer/add
        parameters:
            Name:string     //message name
            URL:string      //URL of consume message
            Timeout:int     // milliseconds waiting for response when access url , usually : 3000
            Code:int        //http code,this code decide the url is accessed success or fail,
                              usually it is 200
            CheckCode:1|0   //whether to check response http code when access url,1:true,0:false
            Comment:string  //comment of consumer
            RouteKey:string //routing key
            Token:string    //should be set when IsNeedToken is 1,other leave empty
            api-token:string//the api token is setting in config
            callback:string //callback function name for jsonp call,if no jsonp call ,leave it empty
    response:
        type:json
        column:
            code:1|0    //1 means success , 0 means fail
        example:
            no jsonp:{code:1,data:null} or {code:0,data:"some error"} 
            jsonp:callbackxxx({code:1,data:null}) or callbackxxx({code:0,data:"some error"}) 
5.update a consumer
    request:
        protocol:http
        method:get
        path:/consumer/update
        parameters:
            Name:string     //message name
            ID:string       //ID of consumer
            URL:string      //URL of consume message
            Timeout:int     //milliseconds waiting for response when access url ,
                               usually : 3000
            Code:int        //http code,this code decide the url is accessed success or fail,
                              usually it is 200
            CheckCode:1|0   //whether to check response http code when access url,1:true,0:false
            Comment:string  //comment of consumer
            RouteKey:string //routing key
            Token:string    //should be set when IsNeedToken is 1,other leave empty
            api-token:string//the api token is setting in config
            callback:string //callback function name for jsonp call,
                              if no jsonp call ,leave it empty
    response:
        type:json
        column:
            code:1|0    //1 means success , 0 means fail
        example:
            no jsonp:{code:1,data:null} or {code:0,data:"some error"} 
            jsonp:callbackxxx({code:1,data:null}) or callbackxxx({code:0,data:"some error"})
6.delete a consumer
    request:
        protocol:http
        method:get
        path:/consumer/delete
        parameters:
            Name:string     //message name
            ID:string       //ID of consumer
            api-token:string//the api token is setting in config
            callback:string //callback function name for jsonp call,
                              if no jsonp call ,leave it empty
    response:
        type:json
        column:
            code:1|0    //1 means success , 0 means fail
        example:
            no jsonp:{code:1,data:null} or {code:0,data:"some error"} 
            jsonp:callbackxxx({code:1,data:null}) or callbackxxx({code:0,data:"some error"}
7.restart service
    request:
            protocol:http
            method:get
            path:/restart
            parameters:
                api-token:string    //the api token is setting in config
                callback:string     //callback function name for jsonp call,
                                      if no jsonp call ,leave it empty
    response:
            type:json
            column:
                code:1|0    //1 means success , 0 means fail
            example:
                no jsonp:{code:1,data:null} or {code:0,data:"some error"} 
                jsonp:callbackxxx({code:1,data:null}) or callbackxxx({code:0,data:"some error"})
8.reload service
    request:
            protocol:http
            method:get
            path:/reload
            parameters:
                api-token:string    //the api token is setting in config
                callback:string     //callback function name for jsonp call,
                                      if no jsonp call ,leave it empty
    response:
            type:json
            column:
                code:1|0    //1 means success , 0 means fail
            example:
                no jsonp:{code:1,data:null} or {code:0,data:"some error"} 
                jsonp:callbackxxx({code:1,data:null}) or callbackxxx({code:0,data:"some error"})
9.get messages config  
    request:
            protocol:http
            method:get
            path:/config
            parameters:
                Name:string         //message name
                api-token:string    //the api token is setting in config
                callback:string     //callback function name for jsonp call,
                                      if no jsonp call ,leave it empty
    response:
            type:json
            column:
                code:1|0    //1 means success , 0 means fail
            example:
                no jsonp:
                            [{
                                "Comment": "",
                                "Consumers": [{
                                        "Comment": "",
                                        "ID": "111",
                                        "Code": 200,
                                        "CheckCode": true,
                                        "RouteKey": "#",
                                        "Timeout": 5000,
                                        "URL": "http://test.com/wmq.php"
                                    }
                                ],
                                "Durable": false,
                                "IsNeedToken": true,
                                "Mode": "topic",
                                "Name": "test",
                                "Token": "JQJsUOqYzYZZgn8gUvs7sIinrJ0tDD8J"
                            }]
                 or {code:0,data:"some error"} 
                jsonp:callbackxxx({code:1,data:[...]}) or callbackxxx({code:0,data:"some error"})
10.get a consumer status
    request:
            protocol:http
            method:get
            path:/consumer/status
            parameters:
                Name:string         //message name
                api-token:string    //the api token is setting in config
                callback:string     //callback function name for jsonp call,
                                      if no jsonp call ,leave it empty
    response:
            type:json
            column:
                code:1|0    //1 means success , 0 means fail
            example:
                no jsonp:
                            {
                                "code": 1, 
                                "data": {
                                    "Count": 0, 
                                    "ID": "111", 
                                    "LastTime": "1496480916", 
                                    "MsgName": "test"
                                }
                            }
                 or {code:0,data:"some error"} 
                jsonp:callbackxxx({code:1,data:[...]}) or callbackxxx({code:0,data:"some error"})
11.get all consumer status of a message
    request:
            protocol:http
            method:get
            path:/message/status
            parameters:
                Name:string         //message name
                ID:string           //consumer's ID
                api-token:string    //the api token is setting in config
                callback:string     //callback function name for jsonp call,
                                      if no jsonp call ,leave it empty
    response:
            type:json
            column:
                code:1|0    //1 means success , 0 means fail
            example:
                no jsonp:
                            {
                                "code": 1, 
                                "data": [
                                    {
                                        "Count": 0, 
                                        "ID": "111", 
                                        "LastTime": "1496480916", 
                                        "MsgName": "test"
                                    }, 
                                    {
                                        "Count": 0, 
                                        "ID": "222", 
                                        "LastTime": "1496480916", 
                                        "MsgName": "test"
                                    }
                                ]
                            }
                 or {code:0,data:"some error"} 
                jsonp:callbackxxx({code:1,data:[...]}) or callbackxxx({code:0,data:"some error"})
12.get or search last 100 lines log content
    request:
            protocol:http
            method:get
            path:/log
            parameters:
                keyword:string           //keyword to search
                type:string             //should be one of: info,error,debug
                api-token:string        //the api token is setting in config
                callback:string         //callback function name for jsonp call,
                                            if no jsonp call ,leave it empty
    response:
            type:json
            column:
                code:1|0    //1 means success , 0 means fail
            example:
                no jsonp:
                            {
                                "code": 1, 
                                "data":"log content"
                            }
                 or {code:0,data:"some error"} 
                jsonp:callbackxxx({code:1,data:[...]}) or callbackxxx({code:0,data:"some error"})
13.get all log file names
    request:
            protocol:http
            method:get
            path:/log/list
            parameters:
                api-token:string        //the api token is setting in config
                callback:string         //callback function name for jsonp call,
                                            if no jsonp call ,leave it empty
    response:
            type:json
            column:
                code:1|0    //1 means success , 0 means fail
            example:
                no jsonp:{"code":1,"data":["error.log","info.log"]}
                 or {code:0,data:"some error"} 
                jsonp:callbackxxx({code:1,data:[...]}) or callbackxxx({code:0,data:"some error"})
14.download a log file
    request:
            protocol:http
            method:get
            path:/log/file
            parameters:
                file:string             //filename of log file
                api-token:string        //the api token is setting in config
    response:
            your browser will tip download file
</pre>
