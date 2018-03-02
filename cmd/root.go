package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/mpppk/kniv/downloader"
	"github.com/mpppk/kniv/kniv"
	_ "github.com/mpppk/kniv/tumblr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "kniv",
	Short: "crawler",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		dl := downloader.New(100000, 3000, viper.GetString("downloader/destination_dir"))

		crawlersSetting := viper.GetStringMap("crawlers")

		for _, crawlerGenerator := range kniv.CrawlerFactories {
			crawler, err := crawlerGenerator.Create(crawlersSetting)
			if err != nil {
				log.Fatal(err)
			}
			dl.RegisterCrawler(crawler)
			crawler.SendResourceUrlsToChannel()
		}
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
		viper.AddConfigPath(path.Join(home, ".config", "kniv"))
		viper.SetConfigName(".kniv")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
