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

var config = &i18n.LocalizeConfig{}

func (s *Server) getMessage(msgId string, langs ...string) string {
	config.MessageID = msgId
	for _, lang := range langs {
		localizer := s.LocalizerMap[lang]
		if localizer != nil {
			return localizer.MustLocalize(config)
		}
	}
	localizer := s.LocalizerMap["zh"]
	return localizer.MustLocalize(config)
}
