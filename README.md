# WMQ
wrapped  message queue which based on rabbitmq,support http protocol
# Usage:
#wmq -h

# Note

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
            Code:int        //http code,this code decide the url is accessed success or fail,usually it is 200
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
10.get all consumer status of a message
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
11.get a consumer status
    request:
            protocol:http
            method:get
            path:/consumer/status
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
        path:/:name?:query_string     //:name is the name of message ,:query_string is any query string you need
        header:
            Token:string        //message's Token , if not need token ,leave it empty
            RouteKey:string     //the api token is setting in config
    response:
        httpcode:204|500      //204:menas success 500:means fail and output is error info
</pre>
