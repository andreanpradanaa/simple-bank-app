package mail

import (
	"testing"

	"github.com/andreanpradanaa/simple-bank-app/utils"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config, err := utils.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "LOVE YOU SAYANGGG. Tiatii dijalannn, semoga makanannyo enakkk <3"
	content := `
		<h1> LOVE YOU SAYANGGG. Tiatii dijalannn, semoga makanannyo enakkk <3 </h1>
		<p> <3 </p>
		`

	to := []string{"andrean.pradana001@gmail.com"}
	cc := []string{"fidyahputri15@gmail.com"}
	attachFiles := []string{"../readme.md"}

	err = sender.SendEmail(subject, content, to, cc, nil, attachFiles)
	require.NoError(t, err)
}
