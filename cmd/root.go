package cmd

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
	"github.com/mpppk/kniv/kniv"
	//_ "github.com/mpppk/kniv/tumblr"
	"log"
	"os"
	"path"
	"time"

	"github.com/mpppk/kniv/twitter"
	_ "github.com/mpppk/kniv/twitter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "kniv",
	Short: "real time event stream engine",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// FIXME don't load config via viper
		crawlersSetting := viper.GetStringMap("processors")
		setting, ok := crawlersSetting["twitter"].(map[string]interface{})
		if !ok {
			log.Fatal("invalid twitter setting")
		}

		flow := kniv.LoadFlowFromFile("sample_flow.yml")
		fmt.Printf("%#v\n", flow)

		dispatcher := kniv.NewDispatcher(100000)

		twitterProcessor, err := twitter.NewProcessorFromConfigMap(100000, setting)
		if err != nil {
			log.Fatal(err)
		}

		//processorGenerators := []kniv.ProcessorGenerator{
		//	&kniv.DelayProcessorGenerator{},
		//}

		factory := kniv.ProcessorFactory{}
		factory.AddGenerator(&kniv.DelayProcessorGenerator{})

		baseArgs := &kniv.BaseArgs{QueueSize: 100000}
		delayArgs := &kniv.DelayProcessorArgs{BaseArgs: baseArgs, IntervalMilliSec: 5000, Group: "test"}

		delayProcessor := kniv.NewDelayProcessor(delayArgs)

		tasks := []kniv.FilterTask{
			kniv.NewFilterByJSTask([]string{"p.downloaded"}),
			kniv.NewDistinctTask([]string{"since_id"}),
			kniv.NewSelectPayloadTask([]string{"since_id", "count", "user", "group"}),
		}
		customProcessor := kniv.NewCustomProcessor(100000, tasks)

		dispatcher.RegisterTask(twitterProcessor.Name, []kniv.Label{"init", "twitter"}, []kniv.Label{"transform", "download", "delay"}, twitterProcessor) // FIXME
		dispatcher.RegisterTask(delayProcessor.Name, []kniv.Label{"delay"}, []kniv.Label{}, delayProcessor)
		dispatcher.RegisterTask("downloader", []kniv.Label{"download"}, []kniv.Label{}, kniv.NewDownloader(100000, "idp"))
		dispatcher.RegisterTask(customProcessor.Name, []kniv.Label{"transform"}, []kniv.Label{"twitter"}, customProcessor)
		go dispatcher.Start()
		dispatcher.StartProcessors()
		initEvent := &kniv.BaseEvent{} // FIXME
		initEvent.SetSourceId(0)
		initEvent.PushLabel("init")
		initEvent.SetPayload(map[string]interface{}{
			"count": 100,
		})
		dispatcher.AddResource(initEvent)
		time.Sleep(10 * time.Minute) // FIXME
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
