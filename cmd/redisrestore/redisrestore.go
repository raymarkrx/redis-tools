package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"github.com/wooos/redis-tools/internal/instance"
	"github.com/wooos/redis-tools/internal/version"
)

type restoreValues struct {
	host     string
	port     int
	password string
	version  bool
}

func main() {
	v := &restoreValues{}

	cmd := &cobra.Command{
		Use:   "redisrestore <FILENAME>",
		Short: "Restore backups generated with redisdump to a running server.",
		Run:   v.RunCommand,
	}

	flags := cmd.Flags()
	flags.StringVar(&v.host, "host", "localhost", "server hostname")
	flags.IntVar(&v.port, "port", 6379, "server port")
	flags.StringVar(&v.password, "password", "", "password to use when connecting to the server")
	flags.BoolVarP(&v.version, "version", "v", false, "print version information")

	_ = cmd.Execute()
}

func (v *restoreValues) RunCommand(cmd *cobra.Command, args []string) {
	if v.version {
		ver := version.GetVersion()
		fmt.Println(ver)
		return
	}

	if len(args) != 1 {
		fmt.Printf("%q requires %d %s\n\nUsage:  %s\n", cmd.CommandPath(), 1, "argument", cmd.UseLine())
		return
	}

	file := args[0]

	ins, err := instance.LoadFromFile(file)
	if err != nil {
		log.Fatalf("Cannot load file: %s, error: %v\n", file, err)
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
	ins.Restore(conn)
}
