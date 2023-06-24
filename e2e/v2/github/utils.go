package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type want struct {
	status int
}

func genSignature(t *testing.T, payload interface{}) string {
	message, err := json.Marshal(payload)
	require.NoError(t, err)

	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write(message)

	return fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
}
