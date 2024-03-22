/*
Copyright © 2024 shynome <shynome@gmail.com>
*/
package cmd

import (
	"context"
	"errors"
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/shynome/err0/try"
	"github.com/spf13/cobra"
	"remoon.net/wslink/client"
)

// cCmd represents the c command
var cCmd = &cobra.Command{
	Use:   "c [ws] [peer] [http server]",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 3 {
			log.Println("参数不完整")
			return
		}

		target := try.To1(url.Parse(args[2]))
		handler := httputil.NewSingleHostReverseProxy(target)
		client := client.New(handler)
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			<-c
			cancel()
		}()
		for {
			sess, err := client.Connect(ctx, args[0], args[1])
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				log.Println("connect wrong", err)
				time.Sleep(time.Second)
				continue
			}
			log.Println("连接成功")
			select {
			case <-sess.CloseChan():
				log.Println("连接断开, 准备重连中")
			case <-ctx.Done():
				sess.Close()
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(cCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
