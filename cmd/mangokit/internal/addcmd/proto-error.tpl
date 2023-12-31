syntax = "proto3";

package {{.FileName}}.v1;

import "errors/errors.proto";

option go_package = "{{.Package}};v1";

enum {{.Name}} {
	option (errors.default_code) = 500;

	Placeholder = 0 [(errors.code) = 0];

}
