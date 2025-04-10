package services

import (
	"fmt"
	"github.com/faelmori/logz"
	"math/rand"
	"os"
	"path/filepath"
)

func splitMessage(recPayload []string) (id, msg []string) {
	if recPayload[1] == "" {
		id = recPayload[:2]
		msg = recPayload[2:]
	} else {
		id = recPayload[:1]
		msg = recPayload[1:]
	}
	return
}
func getBrokersPath() (string, error) {
	brkDir, homeErr := os.UserHomeDir()
	if homeErr != nil || brkDir == "" {
		brkDir, homeErr = os.UserConfigDir()
		if homeErr != nil || brkDir == "" {
			brkDir, homeErr = os.UserCacheDir()
			if homeErr != nil || brkDir == "" {
				brkDir = "/tmp"
			}
		}
	}

	brkDir = filepath.Join(brkDir, ".kubex", "gkbxsrv", "brokers")

	if _, statErr := os.Stat(brkDir); statErr != nil {
		if mkDirErr := os.MkdirAll(brkDir, 0755); mkDirErr != nil {
			logz.ErrorCtx("Error creating brokers", map[string]interface{}{
				"context":  "gkbxsrv",
				"action":   "getBrokerPath",
				"showData": true,
				"error":    mkDirErr.Error(),
			})
			return "", mkDirErr
		}
	}

	logz.InfoCtx(fmt.Sprintf("PID's folder: %s", brkDir), map[string]interface{}{
		"context": "gkbxsrv",
		"action":  "getBrokerPath",
	})

	return brkDir, nil
}
func randomName() string {
	return "broker-" + randStringBytes(5)
}
func randStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
