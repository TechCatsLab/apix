package client

import (
	"testing"
)

func TestGetFile(t *testing.T) {
	client := NewClient(nil)

	_, err := client.GetFile("https://raw.githubusercontent.com/TechCatsLab/apix/master/.gitignore", "test")
	if err != nil {
		t.Errorf("fail to get file: %s", err)
	}
}

func TestGetFileWithProxy(t *testing.T) {
	client, err := NewClientWithProxy("http://127.0.0.1:1087")
	if err != nil {
		t.Errorf("fail to set proxy: %s", err)
	}

	_, err = client.GetFile("https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_120x44dp.png", "test")
	if err != nil {
		t.Errorf("fail to get file: %s", err)
	}
}
