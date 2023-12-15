
import (
	"fmt"

	"github.com/mangohow/mangokit/errors"
)

{{ range .Errors}}
{{ if ne .Comment ""}}{{ .Comment }}{{ end -}}
func Error{{ .CamelName }}(format string, args ...interface{}) errors.Error {
	return errors.New(int32({{ .EnumName }}_{{ .Name }}), {{ .HTTPStatus }}, {{ .EnumName }}_{{ .Name }}.String(), fmt.Sprintf(format, args...))
}

{{ end }}
