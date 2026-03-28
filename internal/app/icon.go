package app

import "encoding/base64"

var trayIcon = mustDecodeBase64("iVBORw0KGgoAAAANSUhEUgAAABIAAAASCAYAAABWzo5XAAAAJUlEQVR42mNgoCP4j4TpbxA2TUPDIIKGjxo0atCgStmDL/eTDQCinWuV3kl4AQAAAABJRU5ErkJggg==")

func mustDecodeBase64(value string) []byte {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(err)
	}

	return data
}
