/*
Copyright © 2024 shynome <shynome@gmail.com>
*/
package cmd

import (
	"log"
	"net"
	"net/http"

	"github.com/docker/go-units"
	"github.com/shynome/err0/try"
	"github.com/spf13/cobra"
	"remoon.net/wslink/server"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, _args []string) {
		size := try.To1(units.FromHumanSize(args.size))
		srv := server.New(int(size))
		l := try.To1(net.Listen("tcp", args.listen))
		defer l.Close()
		log.Println("server is running on ", l.Addr().String())
		http.Serve(l, srv)
	},
}

var args struct {
	listen string
	size   string
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	serveCmd.Flags().StringVar(&args.listen, "listen", ":7799", "server listen addr")
	serveCmd.Flags().StringVar(&args.size, "size", "500M", "otter cache size")
}
