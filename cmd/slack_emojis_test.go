package cmd

import (
	"testing"
	"os"
	"fmt"
)

func TestCheckOutputDir(t *testing.T) {
	testDir := "testdir"
	os.Remove(testDir)

	// ディレクトリがない場合 -> 作成出来る
	if err := checkOutputDir(testDir); err != nil {
		t.Fatal(err)
	}
	if s, err := os.Stat(testDir); err != nil {
		t.Fatal(err)
	} else if !s.IsDir() {
		t.Fatal("testDir is not directory")
	}

	// ディレクトリがある場合 -> エラーにならない
	if err := checkOutputDir(testDir); err != nil {
		t.Fatal(err)
	}

	// 同名のファイルがある場合 -> エラーになる
	os.Remove(testDir)
	file, _ := os.Create(testDir)
	defer file.Close()

	if err := checkOutputDir(testDir); err == nil {
		t.Fatal("no error!!")
	}

	os.Remove(testDir)
}

func TestNewSlackEmojisClient(t *testing.T) {
	client := NewSlackEmojisClient()
	if client == nil {
		t.Fatal("client is nil")
	}
}

func TestSlackEmojisClient_GetEmojiList(t *testing.T) {
	client := NewSlackEmojisClient()
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		t.Fatal("SLACK_TOKEN is not defined")
	}

	// invalid token
	el, err := client.GetEmojiList("aaaaaaaa")
	if el.Ok {
		t.Fatalf("got: %v\nwant: %v", el.Ok, false)
	}

	// ok
	el, err = client.GetEmojiList(token)
	if err != nil {
		t.Fatal(err)
	} else if el == nil {
		t.Fatal("el is nil")
	} else if !el.Ok {
		t.Fatalf("got: %v\nwant: %v", el.Ok, true)
	}

}

func TestSlackEmojisClient_GetEmojiFiles(t *testing.T) {
	client := NewSlackEmojisClient()
	w := os.Stdout

	el := &EmojiList{
		Ok: false,
		Error: "hogehogeError",
	}

	testDir := "testdir"
	if err := checkOutputDir(testDir); err != nil {
		t.Fatal(err)
	}

	err := client.GetEmojiFiles(el, testDir, w)
	if err == nil {
		t.Fatal("got: nil\nwant: error")
	}

	el = &EmojiList{
		Ok: true,
		Emoji: map[string]string {"squirrel": "https://my.slack.com/emoji/squirrel/f35f40c0e0.png"},
	}

	err = client.GetEmojiFiles(el, testDir, w)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Stat(fmt.Sprintf("%s/squirrel.png", testDir))
	if err != nil {
		t.Fatal(err)
	} else if f.IsDir() {
		t.Fatal("got: dir\nwant: file")
	} else if f.Size() < 1 {
		t.Fatal("file size is ZERO!!!")
	}

	os.RemoveAll(testDir)
}