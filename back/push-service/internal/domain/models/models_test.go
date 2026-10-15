package models_test

import (
	"testing"

	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDropboxAccount_Folder_NormalizedPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		folder string
		path   string
	}{
		{"Root", "", "/book.epub"},
		{"Slash", "/", "/book.epub"},
		{"Without slashes", "Inkomi", "/Inkomi/book.epub"},
		{"Trailing slash", "/Inkomi/Books/", "/Inkomi/Books/book.epub"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			account, err := models.NewDropboxAccount("token", tc.folder)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.path, account.Path("book.epub"))
		})
	}
}

func TestNewDropboxAccount_EmptyToken_ReturnsErrDropboxNotLinked(t *testing.T) {
	t.Parallel()

	// Act
	account, err := models.NewDropboxAccount("", "/Inkomi")

	// Assert
	assert.ErrorIs(t, err, models.ErrDropboxNotLinked)
	assert.Nil(t, account)
}

func TestNewSentMessage_Msg_Body(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		msg  jobs.Msg
		body string
	}{
		{"Title", jobs.Msg{Title: "Dune", Filename: "dune.epub"}, "Dune is ready"},
		{"Filename fallback", jobs.Msg{Filename: "dune.epub"}, "dune.epub is ready"},
		{"To cloud", jobs.Msg{Title: "Dune", ToCloud: true}, "Dune was sent to Dropbox"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			message := models.NewSentMessage(tc.msg)

			// Assert
			assert.Equal(t, tc.body, message.Body)
			assert.Equal(t, models.PushTypeSent, message.Data["type"])
		})
	}
}

func TestNewSentMessage_Result_IncludesParentId(t *testing.T) {
	t.Parallel()

	// Act
	message := models.NewSentMessage(jobs.Msg{Id: "r1", Result: &jobs.Result{ParentId: "p1"}})

	// Assert
	assert.Equal(t, "r1", message.Data["id"])
	assert.Equal(t, "p1", message.Data["parentId"])
}
