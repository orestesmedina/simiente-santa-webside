package storage

import "testing"

// Firmas binarias mínimas con las que http.DetectContentType clasifica el
// contenido (T309): no se usa la extensión ni el Content-Type del cliente.
var (
	jpegBytes = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}
	pngBytes  = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x01, 0x02, 0x03}
	webpBytes = []byte{'R', 'I', 'F', 'F', 0x04, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P', 'V', 'P', '8', ' '}
	gifBytes  = []byte("GIF89a\x01\x00\x01\x00")
	svgBytes  = []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`)
	htmlBytes = []byte("<html><body>hola</body></html>")
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantType  string
		wantExt   string
		wantValid bool
	}{
		{name: "jpeg", data: jpegBytes, wantType: "image/jpeg", wantExt: "jpg", wantValid: true},
		{name: "png", data: pngBytes, wantType: "image/png", wantExt: "png", wantValid: true},
		{name: "webp", data: webpBytes, wantType: "image/webp", wantExt: "webp", wantValid: true},
		{name: "gif rechazado", data: gifBytes, wantValid: false},
		{name: "svg rechazado", data: svgBytes, wantValid: false},
		{name: "html renombrado rechazado", data: htmlBytes, wantValid: false},
		{name: "vacío rechazado", data: nil, wantValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mediaType, ext, ok := Detect(tt.data)
			if ok != tt.wantValid {
				t.Fatalf("Detect() ok = %v, se esperaba %v", ok, tt.wantValid)
			}
			if !ok {
				return
			}
			if mediaType != tt.wantType {
				t.Errorf("Detect() mediaType = %q, se esperaba %q", mediaType, tt.wantType)
			}
			if ext != tt.wantExt {
				t.Errorf("Detect() ext = %q, se esperaba %q", ext, tt.wantExt)
			}
		})
	}
}

func TestValidName(t *testing.T) {
	const good = "img_0123456789abcdef0123456789abcdef0123.jpg"
	tests := []struct {
		name string
		want bool
	}{
		{name: good, want: true},
		{name: "img_0123456789abcdef0123456789abcdef0123.webp", want: true},
		{name: "../../etc/passwd", want: false},
		{name: "img_abc.jpg", want: false},
		{name: "/etc/passwd", want: false},
		{name: "img_0123456789abcdef0123456789abcdef0123.svg", want: false},
		{name: "img_0123456789ABCDEF0123456789abcdef0123.png", want: false},
		{name: "img_0123456789abcdef0123456789abcdef0123.gif", want: false},
		{name: "", want: false},
	}
	for _, tt := range tests {
		if got := ValidName(tt.name); got != tt.want {
			t.Errorf("ValidName(%q) = %v, se esperaba %v", tt.name, got, tt.want)
		}
	}
}

func TestContentTypeFor(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "img_0123456789abcdef0123456789abcdef0123.jpg", want: "image/jpeg"},
		{name: "img_0123456789abcdef0123456789abcdef0123.png", want: "image/png"},
		{name: "img_0123456789abcdef0123456789abcdef0123.webp", want: "image/webp"},
		{name: "cualquiera.bin", want: "application/octet-stream"},
		{name: "", want: "application/octet-stream"},
	}
	for _, tt := range tests {
		if got := ContentTypeFor(tt.name); got != tt.want {
			t.Errorf("ContentTypeFor(%q) = %q, se esperaba %q", tt.name, got, tt.want)
		}
	}
}
