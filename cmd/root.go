package cmd

import (
	"fmt"
	"os"
	"strconv"

	"log"

	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	"github.com/mpppk/kniv/downloader"
	"github.com/mpppk/kniv/tumblr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "kniv",
	Short: "crawler",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		offset := 0
		if len(os.Args) > 1 {
			num, err := strconv.Atoi(os.Args[1])
			if err != nil {
				log.Fatal(err)
			}
			offset = num
		}

		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		tumblrCrawler := tumblr.NewCrawler(&tumblr.Opt{
			ConsumerKey:              os.Getenv("CONSUMER_KEY"),
			ConsumerSecret:           os.Getenv("CONSUMER_SECRET"),
			OauthToken:               os.Getenv("OAUTH_TOKEN"),
			OauthSecret:              os.Getenv("OAUTH_SECRET"),
			Offset:                   offset,
			MaxBlogNum:               200,
			PostNumPerBlog:           500,
			APIIntervalMilliSec:      4000,
			DownloadIntervalMilliSec: 3000,
			DstDirMap: map[string]string{
				"photo": "imgs",
				"video": "videos",
			},
		})

		downloader := downloader.New(100000, 3000)
		downloader.RegisterCrawler(tumblrCrawler, "sammple_img")
		tumblrCrawler.SendResourceUrlsToChannel()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kniv.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kniv" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kniv")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
