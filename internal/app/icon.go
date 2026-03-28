package app

import "encoding/base64"

var trayIcon = mustDecodeBase64("iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAA2ElEQVR42u3aQQ6EMAxDUe5/aXMDNNBOaiffUpeQ5kFBgl4XIYScjaY0+XWMazgaRIVjbONWEDIaEc07nf/IlXerYf1Ks1sO1eu05XMhFkChANp5VVMBtOuWTgZQR4B/zP9xAukAWi3eAUAALBTuAiAAAPhWtBOAAAAAAAAAAKC2oBuAuAMAAAAAAAAAgO8BAADAV2EAfjm4A8DrjP436A5Q/os8GWA5o3eIVK/bnQClzY3ZJTZ2n+COybnVKEdwOr/Nc4Ft88mNn4SwztjGI19pDjiEkDO5AXah/khxODX9AAAAAElFTkSuQmCC")

func mustDecodeBase64(value string) []byte {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(err)
	}

	return data
}
