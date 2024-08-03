syntax = "proto3";

package {{.FileName}};

import "google/api/annotations.proto";

option go_package = "{{.Package}};{{.DirName}}";

service {{.Name}} {

}
