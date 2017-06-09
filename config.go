package main

import (
	"flag"
	"fmt"

	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfg = viper.New()
)

func initConfig() (err error) {
	cfg.SetDefault("wmq.version", "1.0")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.String("listen-api", "0.0.0.0:3302", "api service listening port")
	pflag.String("listen-publish", "0.0.0.0:3303", "publish service listening port")
	pflag.String("api-token", "guest", "access api token")
	configFile := pflag.String("config", "", "config file path")
	pflag.Bool("api-disable", false, "disable api service")
	pflag.String("level", "debug", "console log level,should be one of debug,info,warn,error")
	version := pflag.Bool("version", false, "show version about current WMQ")
	example := pflag.Bool("data-example", false, "print example of data-file")
	pflag.StringSlice("ignore-headers", []string{}, "these http headers will be ignored when access to consumer's url , multiple splitted by comma(,)")
	pflag.String("realip-header", "X-Forwarded-For", "the publisher's real ip will be set in this http header when access to consumer's url")
	pflag.Int("fail-wait", 50000, "access consumer url  fail and then how many milliseconds to sleep")
	pflag.String("mq-host", "127.0.0.1", "which host be used when connect to RabbitMQ")
	pflag.Int("mq-port", 5672, "which port be used when connect to RabbitMQ")
	pflag.String("mq-username", "guest", "which username be used when connect to RabbitMQ")
	pflag.String("mq-password", "guest", "which password be used when connect to RabbitMQ")
	pflag.String("mq-vhost", "/", "which vhost be used when connect to RabbitMQ")
	pflag.String("mq-prefix", "wmq.", "the queue and exchange default prefix")
	pflag.String("data-file", "message.json", "which file will store messages")
	pflag.String("log-dir", "log", "the directory which store log files")
	pflag.Bool("log-access", true, "access log on or off")
	pflag.Int64("log-max-size", 102400000, "log file max size(bytes) for rotate")
	pflag.Int("log-max-count", 3, "log file max count for rotate to remain")
	pflag.StringSlice("log-level", []string{"info", "error", "debug"}, "log to file level,multiple splitted by comma(,)")
	pflag.Parse()
	if *version {
		fmt.Printf("WMQ v%s - https://github.com/snail007/wmq\n", cfg.GetString("wmq.version"))
		os.Exit(0)
	}
	if *example {
		printExample()
		os.Exit(0)
	}
	cfg.BindPFlag("listen.api", pflag.Lookup("listen-api"))
	cfg.BindPFlag("listen.publish", pflag.Lookup("listen-publish"))
	cfg.BindPFlag("api.token", pflag.Lookup("api-token"))
	cfg.BindPFlag("api.disable", pflag.Lookup("api-disable"))
	cfg.BindPFlag("publish.IgnoreHeaders", pflag.Lookup("ignore-headers"))
	cfg.BindPFlag("publish.RealIpHeader", pflag.Lookup("realip-header"))
	cfg.BindPFlag("consume.FailWait", pflag.Lookup("fail-wait"))
	cfg.BindPFlag("consume.DataFile", pflag.Lookup("data-file"))
	cfg.BindPFlag("rabbitmq.host", pflag.Lookup("mq-host"))
	cfg.BindPFlag("rabbitmq.port", pflag.Lookup("mq-port"))
	cfg.BindPFlag("rabbitmq.username", pflag.Lookup("mq-username"))
	cfg.BindPFlag("rabbitmq.password", pflag.Lookup("mq-password"))
	cfg.BindPFlag("rabbitmq.vhost", pflag.Lookup("mq-vhost"))
	cfg.BindPFlag("rabbitmq.prefix", pflag.Lookup("mq-prefix"))
	cfg.BindPFlag("log.dir", pflag.Lookup("log-dir"))
	cfg.BindPFlag("log.level", pflag.Lookup("log-level"))
	cfg.BindPFlag("log.access", pflag.Lookup("log-access"))
	cfg.BindPFlag("log.console-level", pflag.Lookup("level"))
	cfg.BindPFlag("log.fileMaxSize", pflag.Lookup("log-max-size"))
	cfg.BindPFlag("log.maxCount", pflag.Lookup("log-max-count"))
	cfg.SetDefault("default.IgnoreHeaders", []string{"Token", "RouteKey", "Host", "Accept-Encoding", " Content-Length", "Content-Type Connection"})
	fmt.Printf("%s", *configFile)
	if *configFile != "" {
		cfg.SetConfigFile(*configFile)
	} else {
		cfg.SetConfigName("config")
		cfg.AddConfigPath("/etc/wmq/")
		cfg.AddConfigPath("$HOME/.wmq")
		cfg.AddConfigPath(".wmq")
		cfg.AddConfigPath(".")
	}
	err = cfg.ReadInConfig()

	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("%s", err)
	} else {
		fmt.Printf("use config file : %s\n", cfg.ConfigFileUsed())
	}
	cfg.Set("publish.IgnoreHeaders", append(cfg.GetStringSlice("default.IgnoreHeaders"), cfg.GetStringSlice("publish.IgnoreHeaders")...))
	return
}

func printExample() {
	fmt.Println(`[{
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
}]`)
}
