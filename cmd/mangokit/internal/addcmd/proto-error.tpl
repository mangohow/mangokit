syntax = "proto3";

package {{.FileName}};

import "errors/errors.proto";

option go_package = "{{.Package}};{{.DirName}}";

enum {{.Name}} {
	option (errors.default_code) = 500;

	Placeholder = 0 [(errors.code) = 0];

}
