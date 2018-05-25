package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

var cfgFile string

var RootCmd = newRootCmd()

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.slack-emojis.yml)")

	viper.BindPFlag("url", RootCmd.PersistentFlags().Lookup("url"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".slack-emojis")
	viper.AddConfigPath("$HOME")
	viper.AutomaticEnv()

	viper.ReadInConfig()
}

func newRootCmd() *cobra.Command {
	type Options struct {
		OptToken     string
		OptOutputDir string
	}

	var (
		o = &Options{}
	)

	type EmojiList struct {
		Ok      bool              `json:"ok"`
		Emoji   map[string]string `json:"emoji"`
		CacheTs string            `json:"cache_ts"`
		Error   string            `json:"error"`
	}

	cmd := &cobra.Command{
		Use:   "slack-emojis",
		Short: "Slack Emoji Downloader",
		RunE: func(cmd *cobra.Command, args []string) error {

			if _, err := os.Stat(o.OptOutputDir); err != nil {
				if err := os.Mkdir(o.OptOutputDir, 0755); err != nil {
					return err
				}
			}
			if s, _ := os.Stat(o.OptOutputDir); !s.IsDir() {
				return fmt.Errorf("error: output is File: %s", o.OptOutputDir)
			}

			client := new(http.Client)

			url := fmt.Sprintf("https://slack.com/api/emoji.list?token=%s", o.OptToken)

			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			res, _ := client.Do(req)
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return err
			}

			el := &EmojiList{}
			if err := json.Unmarshal(body, el); err != nil {
				return err
			}

			if el.Ok {
				var res2 *http.Response
				var file *os.File

				for key, value := range el.Emoji {
					if !strings.HasPrefix(value, "alias:") {
						req2, _ := http.NewRequest("GET", value, nil)
						res2, _ = client.Do(req2)

						filename := fmt.Sprintf("%s%s", key, value[strings.LastIndex(value, "."):])
						file, err = os.Create(fmt.Sprintf("%s/%s", o.OptOutputDir, filename))
						if err != nil {
							return err
						}

						io.Copy(file, res2.Body)
						cmd.Printf("download: %s\n", filename)
						time.Sleep(1 * time.Second)
					}
				}
				defer res2.Body.Close()
				defer file.Close()
			}

			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.Flags().StringVarP(&o.OptToken, "token", "t", "", "slack authentication token [required]")
	cmd.Flags().StringVarP(&o.OptOutputDir, "output", "o", "./", "emoji output directory (default current directory)")

	cmd.MarkFlagRequired("token")

	return cmd
}
