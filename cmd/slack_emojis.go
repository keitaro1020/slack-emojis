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

var Cmd = SlackEmojisCmd()

func init() {
	cobra.OnInitialize(initConfig)

	Cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.slack-emojis.yml)")

	viper.BindPFlag("url", Cmd.PersistentFlags().Lookup("url"))
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

type Options struct {
	OptToken     string
	OptOutputDir string
}

type EmojiList struct {
	Ok      bool              `json:"ok"`
	Emoji   map[string]string `json:"emoji"`
	CacheTs string            `json:"cache_ts"`
	Error   string            `json:"error"`
}

type SlackEmojisClient struct {
	client *http.Client
}

func NewSlackEmojisClient() *SlackEmojisClient {
	return &SlackEmojisClient{
		client: new(http.Client),
	}
}

var (
	o = &Options{}
)

func SlackEmojisCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:           "slack-emojis",
		Short:         "Slack Emoji Downloader",
		RunE:          slackEmojiFunction,
		SilenceErrors: true,
		SilenceUsage:  false,
	}
	cmd.Flags().StringVarP(&o.OptToken, "token", "t", "", "slack authentication token [required]")
	cmd.Flags().StringVarP(&o.OptOutputDir, "output", "o", "./", "emoji output directory (default current directory)")

	cmd.MarkFlagRequired("token")

	return cmd
}

func slackEmojiFunction(cmd *cobra.Command, args []string) error {
	if err := checkOutputDir(o.OptOutputDir); err != nil {
		return err
	}

	c := NewSlackEmojisClient()

	el, err := c.GetEmojiList(o.OptToken)
	if err != nil {
		return err
	}

	c.GetEmojiFiles(el, o.OptOutputDir, cmd.OutOrStdout())

	cmd.OutOrStdout()

	return nil
}

func checkOutputDir(outputDir string) error {
	if _, err := os.Stat(outputDir); err != nil {
		if err := os.Mkdir(outputDir, 0755); err != nil {
			return err
		}
	}
	if s, _ := os.Stat(outputDir); !s.IsDir() {
		return fmt.Errorf("error: output is File: %s", o.OptOutputDir)
	}
	return nil
}

func (c *SlackEmojisClient) GetEmojiList(token string) (*EmojiList, error) {
	url := fmt.Sprintf("https://slack.com/api/emoji.list?token=%s", token)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, _ := c.client.Do(req)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	el := &EmojiList{}
	if err := json.Unmarshal(body, el); err != nil {
		return nil, err
	}
	return el, nil
}

func (c *SlackEmojisClient) GetEmojiFiles(el *EmojiList, outputDir string, w io.Writer) error {
	if el.Ok {
		var res2 *http.Response
		var file *os.File

		for key, value := range el.Emoji {
			if !strings.HasPrefix(value, "alias:") {
				req2, _ := http.NewRequest("GET", value, nil)
				res2, _ = c.client.Do(req2)

				filename := fmt.Sprintf("%s%s", key, value[strings.LastIndex(value, "."):])
				file, err := os.Create(fmt.Sprintf("%s/%s", outputDir, filename))
				if err != nil {
					return err
				}

				io.Copy(file, res2.Body)
				fmt.Fprintf(w, "download: %s\n", filename)
				time.Sleep(1 * time.Second)
			}
		}
		defer res2.Body.Close()
		defer file.Close()
	} else {
		return fmt.Errorf("error: EmojiList error: %s", el.Error)
	}
	return nil
}
