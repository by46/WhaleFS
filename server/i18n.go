package server

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	MsgIdBucketNameNotCorrect = "bucketNameNotCorrect"
	MsgIdInvalidParam         = "invalidParam"
	MsgIdFileNotFound         = "fileNotFound"
	MsgIdFileTooLarge         = "fileTooLarge"
	MsgIdNoFileContent        = "noFileContent"
	MsgIdFileUrlCannotBeEmpty = "fileUrlCannotBeEmpty"
	MsgIdParamParseFailed     = "paramParseFailed"
	MsgIdPdfFilePathNotSet    = "pdfFilePathNotSet"
	MsgIdMergePdfFailed       = "mergePdfFailed"
	MsgIdMissFileIdentity     = "missFileIdentity"
	MsgIdStartPositionNotSet  = "startPositionNotSet"
)

func (s *Server) getMessage(msgId string, langs ...string) string {
	localizer := i18n.NewLocalizer(s.I18nBundle, langs...)
	return localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: msgId,
	})
}
