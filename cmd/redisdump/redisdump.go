package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"github.com/wooos/redis-tools/internal/instance"
	"github.com/wooos/redis-tools/internal/resolver"
	"github.com/wooos/redis-tools/internal/version"
)

type dumpValues struct {
	host     string
	port     int
	password string
	out      string
	version  bool
}

func main() {
	v := &dumpValues{}

	cmd := &cobra.Command{
		Use:          "redisdump",
		Short:        "Export the content of a running server into .json files.",
		Run:          v.RunCommand,
		SilenceUsage: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&v.host, "host", "localhost", "server hostname")
	flags.IntVar(&v.port, "port", 6379, "server port")
	flags.StringVar(&v.password, "password", "", "password to use when connecting to the server")
	flags.StringVar(&v.out, "out", "redis.json", "output file")
	flags.BoolVarP(&v.version, "version", "v", false, "print version information")

	_ = cmd.Execute()
}

func (v *dumpValues) RunCommand(cmd *cobra.Command, args []string) {
	if v.version {
		ver := version.GetVersion()
		fmt.Println(ver)
		return
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", v.host, v.port),
		Password: v.password,
	})

	ctx, _ := context.WithTimeout(context.TODO(), time.Second*3)

	if _, err := client.Ping(ctx).Result(); err != nil {
		log.Fatalf("Cannot connect to the server, error: %v\n", err)
	}

	conn := client.Conn(ctx)
	keyspace, err := conn.Info(ctx, "keyspace").Result()
	if err != nil {
		log.Fatalf("Cannot get keyspace info, error: %v\n", err)
	}

	dbs := resolver.NewKeySpaces(keyspace)

	ins, err := instance.LoadFromRedis(conn, dbs)
	if err != nil {
		log.Fatalf("Cannot load from redis, error: %v\n", err)
	}

	if err := ins.Dump(v.out); err != nil {
		log.Fatalf("Cannot dump, error: %v\n", err)
	}
}
